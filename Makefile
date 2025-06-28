PACKAGES=$(shell go list ./...)
OUTPUT?=build/fluentumd

BUILD_TAGS?=tendermint,badgerdb

# If building a release, please checkout the version tag to get the correct version setting
ifneq ($(shell git symbolic-ref -q --short HEAD),)
VERSION := unreleased-$(shell git symbolic-ref -q --short HEAD)-$(shell git rev-parse HEAD)
else
VERSION := $(shell git describe)
endif

LD_FLAGS = -X github.com/tendermint/tendermint/version.TMCoreSemVer=$(VERSION)
BUILD_FLAGS = -mod=readonly -ldflags "$(LD_FLAGS)"
HTTPS_GIT := https://github.com/tendermint/tendermint.git
CGO_ENABLED ?= 0

# handle nostrip
ifeq (,$(findstring nostrip,$(TENDERMINT_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
  LD_FLAGS += -s -w
endif

# handle race
ifeq (race,$(findstring race,$(TENDERMINT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_FLAGS += -race
endif

# handle cleveldb
ifeq (cleveldb,$(findstring cleveldb,$(TENDERMINT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += cleveldb
endif

# handle badgerdb
ifeq (badgerdb,$(findstring badgerdb,$(TENDERMINT_BUILD_OPTIONS)))
  BUILD_TAGS += badgerdb
endif

# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(TENDERMINT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += rocksdb
endif

# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(TENDERMINT_BUILD_OPTIONS)))
  BUILD_TAGS += boltdb
endif

# allow users to pass additional flags via the conventional LDFLAGS variable
LD_FLAGS += $(LDFLAGS)

all: check build test install features
.PHONY: all

include tests.mk

###############################################################################
###                              Dependencies                               ###
###############################################################################

# Ensure dependencies are properly managed
deps:
	@echo "--> Ensuring dependencies are up to date"
	@go mod download
	@go mod tidy
	@go mod verify
.PHONY: deps

# Quick dependency check (doesn't modify files)
deps-check:
	@echo "--> Checking dependencies"
	@go mod verify
.PHONY: deps-check

###############################################################################
###                                Build Fluentum                          ###
###############################################################################

# Build with automatic dependency management
build: deps
	@echo "--> Building Fluentum Core $(VERSION)"
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o $(OUTPUT) ./cmd/fluentum/
.PHONY: build

# Build without dependency management (for CI/CD when deps are already managed)
build-only:
	@echo "--> Building Fluentum Core $(VERSION) (dependencies assumed to be ready)"
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o $(OUTPUT) ./cmd/fluentum/
.PHONY: build-only

install: deps
	CGO_ENABLED=$(CGO_ENABLED) go install $(BUILD_FLAGS) -tags $(BUILD_TAGS) ./cmd/fluentum
.PHONY: install

# Install without dependency management
install-only:
	CGO_ENABLED=$(CGO_ENABLED) go install $(BUILD_FLAGS) -tags $(BUILD_TAGS) ./cmd/fluentum
.PHONY: install-only

###############################################################################
###                                Mocks                                    ###
###############################################################################

mockery:
	go generate -run="./scripts/mockery_generate.sh" ./...
.PHONY: mockery

###############################################################################
###                                Protobuf                                 ###
###############################################################################

check-proto-deps:
ifeq (,$(shell which protoc-gen-gogofaster))
	@go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest
endif
.PHONY: check-proto-deps

check-proto-format-deps:
ifeq (,$(shell which clang-format))
	$(error "clang-format is required for Protobuf formatting. See instructions for your platform on how to install it.")
endif
.PHONY: check-proto-format-deps

proto-gen: check-proto-deps
	@echo "Generating Protobuf files"
	@go run github.com/bufbuild/buf/cmd/buf generate
	@mv ./proto/tendermint/abci/types.pb.go ./abci/types/
.PHONY: proto-gen

# These targets are provided for convenience and are intended for local
# execution only.
proto-lint: check-proto-deps
	@echo "Linting Protobuf files"
	@go run github.com/bufbuild/buf/cmd/buf lint
.PHONY: proto-lint

proto-format: check-proto-format-deps
	@echo "Formatting Protobuf files"
	@find . -name '*.proto' -path "./proto/*" -exec clang-format -i {} \;
.PHONY: proto-format

proto-check-breaking: check-proto-deps
	@echo "Checking for breaking changes in Protobuf files against local branch"
	@echo "Note: This is only useful if your changes have not yet been committed."
	@echo "      Otherwise read up on buf's \"breaking\" command usage:"
	@echo "      https://docs.buf.build/breaking/usage"
	@go run github.com/bufbuild/buf/cmd/buf breaking --against ".git"
.PHONY: proto-check-breaking

proto-check-breaking-ci:
	@go run github.com/bufbuild/buf/cmd/buf breaking --against $(HTTPS_GIT)#branch=v0.34.x
.PHONY: proto-check-breaking-ci

###############################################################################
###                              Build ABCI                                 ###
###############################################################################

build_abci: deps
	@go build -mod=readonly -i ./abci/cmd/...
.PHONY: build_abci

install_abci: deps
	@go install -mod=readonly ./abci/cmd/...
.PHONY: install_abci

###############################################################################
###                              Distribution                               ###
###############################################################################

# dist builds binaries for all platforms and packages them for distribution
# TODO add abci to these scripts
dist: deps
	@BUILD_TAGS=$(BUILD_TAGS) sh -c "'$(CURDIR)/scripts/dist.sh'"
.PHONY: dist

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
.PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

draw_deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i github.com/tendermint/tendermint/cmd/tendermint -d 3 | dot -Tpng -o dependency-graph.png
.PHONY: draw_deps

get_deps_bin_size:
	@# Copy of build recipe with additional flags to perform binary size analysis
	$(eval $(shell go build -work -a $(BUILD_FLAGS) -tags $(BUILD_TAGS) -o $(OUTPUT) ./cmd/fluentum/ 2>&1))
	@find $(WORK) -type f -name "*.a" | xargs -I{} du -hxs "{}" | sort -rh | sed -e s:${WORK}/::g > deps_bin_size.log
	@echo "Results can be found here: $(CURDIR)/deps_bin_size.log"
.PHONY: get_deps_bin_size

###############################################################################
###                                  Libs                                   ###
###############################################################################

# generates certificates for TLS testing in remotedb and RPC server
gen_certs: clean_certs
	certstrap init --common-name "tendermint.com" --passphrase ""
	certstrap request-cert --common-name "server" -ip "127.0.0.1" --passphrase ""
	certstrap sign "server" --CA "tendermint.com" --passphrase ""
	mv out/server.crt rpc/jsonrpc/server/test.crt
	mv out/server.key rpc/jsonrpc/server/test.key
	rm -rf out
.PHONY: gen_certs

# deletes generated certificates
clean_certs:
	rm -f rpc/jsonrpc/server/test.crt
	rm -f rpc/jsonrpc/server/test.key
.PHONY: clean_certs

###############################################################################
###                  Formatting, linting, and vetting                       ###
###############################################################################

format:
	find . -name '*.go' -type f -not -path "*.git*" -not -name '*.pb.go' -not -name '*pb_test.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "*.git*"  -not -name '*.pb.go' -not -name '*pb_test.go' | xargs goimports -w -local github.com/tendermint/tendermint
.PHONY: format

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run
.PHONY: lint

DESTINATION = ./index.html.md

###############################################################################
###                           Documentation                                 ###
###############################################################################

build-docs:
	@cd docs && \
	while read -r branch path_prefix; do \
		(git checkout $${branch} && npm ci && VUEPRESS_BASE="/$${path_prefix}/" npm run build) ; \
		mkdir -p ~/output/$${path_prefix} ; \
		cp -r .vuepress/dist/* ~/output/$${path_prefix}/ ; \
		cp ~/output/$${path_prefix}/index.html ~/output ; \
	done < versions ;
.PHONY: build-docs

sync-docs:
	cd ~/output && \
	echo "role_arn = ${DEPLOYMENT_ROLE_ARN}" >> /root/.aws/config ; \
	echo "CI job = ${CIRCLE_BUILD_URL}" >> version.html ; \
	aws s3 sync . s3://${WEBSITE_BUCKET} --profile terraform --delete ; \
	aws cloudfront create-invalidation --distribution-id ${CF_DISTRIBUTION_ID} --profile terraform --path "/*" ;
.PHONY: sync-docs

###############################################################################
###                            Docker image                                 ###
###############################################################################

build-docker: build-linux
	cp $(OUTPUT) DOCKER/tendermint
	docker build --label=tendermint --tag="tendermint/tendermint" DOCKER
	rm -rf DOCKER/tendermint
.PHONY: build-docker

###############################################################################
###                       Local testnet using docker                        ###
###############################################################################

# Build linux binary on other platforms
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build
.PHONY: build-linux

build-docker-localnode:
	@cd networks/local && make
.PHONY: build-docker-localnode

# Runs `make build TENDERMINT_BUILD_OPTIONS=cleveldb` from within an Amazon
# Linux (v2)-based Docker build container in order to build an Amazon
# Linux-compatible binary. Produces a compatible binary at ./build/tendermint
build_c-amazonlinux:
	$(MAKE) -C ./DOCKER build_amazonlinux_buildimage
	docker run --rm -it -v `pwd`:/tendermint tendermint/tendermint:build_c-amazonlinux
.PHONY: build_c-amazonlinux

# Run a 4-node testnet locally
localnet-start: localnet-stop build-docker-localnode
	@if ! [ -f build/node0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/tendermint:Z tendermint/localnode testnet --config /etc/tendermint/config-template.toml --o . --starting-ip-address 192.167.10.2; fi
	docker-compose up
.PHONY: localnet-start

# Stop testnet
localnet-stop:
	docker-compose down
.PHONY: localnet-stop

# Build hooks for dredd, to skip or add information on some steps
build-contract-tests-hooks:
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests.exe ./cmd/contract_tests
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests ./cmd/contract_tests
endif
.PHONY: build-contract-tests-hooks

# Run a nodejs tool to test endpoints against a localnet
# The command takes care of starting and stopping the network
# prerequisits: build-contract-tests-hooks build-linux
# the two build commands were not added to let this command run from generic containers or machines.
# The binaries should be built beforehand
contract-tests:
	dredd
.PHONY: contract-tests

# Fluentum Core Makefile

# Variables
VERSION := v0.1.0
BUILD_DIR := build
BINARY_NAME := fluentumd
MAIN_PACKAGE := ./cmd/fluentum
LDFLAGS := -X github.com/kellyadamtan/tendermint/version.Version=$(VERSION)

# Feature directories
FEATURES_DIR := fluentum/features
QUANTUM_SIGNING_DIR := $(FEATURES_DIR)/quantum_signing
STATE_SYNC_DIR := $(FEATURES_DIR)/state_sync
ZK_ROLLUP_DIR := $(FEATURES_DIR)/zk_rollup

# Quantum Feature Build System Integration
QUANTUM_FEATURE := $(QUANTUM_SIGNING_DIR)/quantum.so

# Default target
.PHONY: all
all: clean build features

# Build the binary
.PHONY: build
build:
	@echo "Building Fluentum Core $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build -tags=plugins -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "$(LDFLAGS)" $(MAIN_PACKAGE)

# Build quantum signing module
.PHONY: build-quantum
build-quantum:
	@echo "Building quantum signing module..."
	@cd $(QUANTUM_SIGNING_DIR) && \
		go build -tags=plugin -buildmode=plugin -o ../quantum.so .

# Install quantum feature
.PHONY: install-quantum
install-quantum: build-quantum
	cp $(QUANTUM_FEATURE) $(GOPATH)/bin/

# Build all features
.PHONY: features
features: feature-quantum-signing feature-state-sync feature-zk-rollup
	@echo "All features built successfully!"

# Build quantum signing feature
.PHONY: feature-quantum-signing
feature-quantum-signing:
	@echo "Building Quantum Signing Feature..."
	@cd $(QUANTUM_SIGNING_DIR) && chmod +x build.sh && ./build.sh
	@echo "Quantum Signing Feature built successfully!"

# Build state sync feature
.PHONY: feature-state-sync
feature-state-sync:
	@echo "Building State Sync Feature..."
	@cd $(STATE_SYNC_DIR) && chmod +x build.sh && ./build.sh
	@echo "State Sync Feature built successfully!"

# Build ZK rollup feature
.PHONY: feature-zk-rollup
feature-zk-rollup:
	@echo "Building ZK Rollup Feature..."
	@cd $(ZK_ROLLUP_DIR) && chmod +x build.sh && ./build.sh
	@echo "ZK Rollup Feature built successfully!"

# Build specific feature
.PHONY: feature
feature:
	@if [ -z "$(FEATURE)" ]; then \
		echo "Error: FEATURE variable not set. Usage: make feature FEATURE=quantum_signing"; \
		exit 1; \
	fi
	@echo "Building $(FEATURE) feature..."
	@cd $(FEATURES_DIR)/$(FEATURE) && chmod +x build.sh && ./build.sh
	@echo "$(FEATURE) feature built successfully!"

# Test all features
.PHONY: test-features
test-features:
	@echo "Testing all features..."
	@cd $(QUANTUM_SIGNING_DIR) && go test -v ./...
	@cd $(STATE_SYNC_DIR) && go test -v ./...
	@cd $(ZK_ROLLUP_DIR) && go test -v ./...
	@echo "All feature tests completed!"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@if exist build rmdir /s /q build
	@if exist .fluentum rmdir /s /q .fluentum

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# Generate protobuf files
.PHONY: proto
proto:
	@echo "Generating protobuf files..."
	@cd proto && buf generate

# Help target
.PHONY: help
help:
	@echo "Fluentum Core - Hybrid Consensus Blockchain Platform"
	@echo ""
	@echo "Dependency Management:"
	@echo "  deps                     - Download and tidy dependencies (auto-run by build)"
	@echo "  deps-check               - Check dependencies without modifying files"
	@echo ""
	@echo "Build Targets:"
	@echo "  build                    - Build binary with automatic dependency management"
	@echo "  build-only               - Build binary (assumes deps are ready)"
	@echo "  install                  - Install binary with automatic dependency management"
	@echo "  install-only             - Install binary (assumes deps are ready)"
	@echo "  clean                    - Clean build artifacts"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  test                     - Run tests"
	@echo "  test-coverage            - Run tests with coverage"
	@echo "  lint                     - Lint the code"
	@echo "  fmt                      - Format the code"
	@echo "  proto-gen                - Generate protobuf code"
	@echo ""
	@echo "Node Management:"
	@echo "  init-node                - Initialize a new node"
	@echo "  start                    - Start the node"
	@echo "  start-dev                - Start the node in development mode"
	@echo "  reset                    - Reset the node"
	@echo "  status                   - Show node status"
	@echo ""
	@echo "Testnet Management:"
	@echo "  init-testnet             - Initialize testnet node"
	@echo "  start-testnet            - Start testnet node"
	@echo "  start-testnet-bg         - Start testnet node in background"
	@echo "  stop-testnet             - Stop testnet node"
	@echo "  testnet-logs             - Show testnet logs"
	@echo "  reset-testnet            - Reset testnet node"
	@echo "  testnet-genesis-account name=NAME - Create testnet genesis account"
	@echo "  testnet-script           - Run testnet startup script (Linux/macOS)"
	@echo "  testnet-script-win       - Run testnet startup script (Windows)"
	@echo ""
	@echo "Account & Key Management:"
	@echo "  keys-add name=NAME       - Add a new account"
	@echo "  keys-list                - List accounts"
	@echo "  keys-show name=NAME      - Show account address"
	@echo ""
	@echo "Genesis & Chain Management:"
	@echo "  add-genesis-account address=ADDR coins=COINS - Add genesis account"
	@echo "  collect-gentxs           - Collect genesis transactions"
	@echo "  validate-genesis         - Validate genesis"
	@echo "  export                   - Export app state"
	@echo ""
	@echo "Query & Transaction Commands:"
	@echo "  query-balance address=ADDR - Query account balance"
	@echo "  tx-send from=FROM to=TO amount=AMT chain-id=CHAIN - Send tokens"
	@echo "  tx-create-fluentum index=IDX title=TITLE body=BODY chain-id=CHAIN - Create Fluentum record"
	@echo "  query-list-fluentum      - List Fluentum records"
	@echo "  query-show-fluentum index=IDX - Show Fluentum record"
	@echo "  query-fluentum-params    - Query Fluentum parameters"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build             - Build Docker image"
	@echo "  docker-run               - Run Docker container"
	@echo "  docker-stop              - Stop Docker container"
	@echo ""
	@echo "Development Notes:"
	@echo "  - 'build' automatically runs 'deps' to ensure dependencies are ready"
	@echo "  - Use 'build-only' in CI/CD when dependencies are pre-managed"
	@echo "  - Run 'deps' manually if you need to update dependencies"
	@echo "  - Use 'deps-check' to verify dependencies without changes"

# Makefile for Fluentum Cosmos SDK Integration

# Variables
BINARY_NAME=fluentumd
BUILD_DIR=build
GO=go
DOCKER=docker

# Build flags
LDFLAGS=-ldflags "-X github.com/fluentum-chain/fluentum/version.Name=Fluentum \
	-X github.com/fluentum-chain/fluentum/version.ServerName=fluentumd \
	-X github.com/fluentum-chain/fluentum/version.ClientName=fluentumcli \
	-X github.com/fluentum-chain/fluentum/version.Version=$(shell git describe --tags --always --dirty) \
	-X github.com/fluentum-chain/fluentum/version.Commit=$(shell git log -1 --format='%H')"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/fluentum

# Install the binary
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) ./cmd/fluentum

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@if exist build rmdir /s /q build
	@if exist .fluentum rmdir /s /q .fluentum

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Lint the code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Generate protobuf code
.PHONY: proto-gen
proto-gen:
	@echo "Generating protobuf code..."
	buf generate

# Initialize a new node
.PHONY: init-node
init-node:
	@echo "Initializing new node..."
	./$(BUILD_DIR)/$(BINARY_NAME) init mynode --chain-id fluentum-local

# Start the node
.PHONY: start
start:
	@echo "Starting node..."
	./$(BUILD_DIR)/$(BINARY_NAME) start

# Start the node in development mode
.PHONY: start-dev
start-dev:
	@echo "Starting node in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) start --log_level debug

# Initialize testnet node
.PHONY: init-testnet
init-testnet:
	@echo "Initializing testnet node..."
	./$(BUILD_DIR)/$(BINARY_NAME) init fluentum-testnet --chain-id fluentum-testnet-1

# Start testnet node
.PHONY: start-testnet
start-testnet:
	@echo "Starting testnet node..."
	./$(BUILD_DIR)/$(BINARY_NAME) start --testnet --api --grpc --grpc-web

# Start testnet node in background
.PHONY: start-testnet-bg
start-testnet-bg:
	@echo "Starting testnet node in background..."
	@nohup ./$(BUILD_DIR)/$(BINARY_NAME) start --testnet --api --grpc --grpc-web > fluentum-testnet.log 2>&1 &
	@echo $$! > fluentum-testnet.pid
	@echo "Node started with PID: $$(cat fluentum-testnet.pid)"

# Stop testnet node
.PHONY: stop-testnet
stop-testnet:
	@echo "Stopping testnet node..."
	@if [ -f fluentum-testnet.pid ]; then \
		kill $$(cat fluentum-testnet.pid) 2>/dev/null || true; \
		rm -f fluentum-testnet.pid; \
		echo "Testnet node stopped"; \
	else \
		echo "No testnet PID file found"; \
	fi

# Show testnet logs
.PHONY: testnet-logs
testnet-logs:
	@echo "Showing testnet logs..."
	@tail -f fluentum-testnet.log

# Reset testnet node
.PHONY: reset-testnet
reset-testnet:
	@echo "Resetting testnet node..."
	./$(BUILD_DIR)/$(BINARY_NAME) tendermint unsafe-reset-all --home ~/.fluentum

# Create testnet genesis account
.PHONY: testnet-genesis-account
testnet-genesis-account:
	@echo "Creating testnet genesis account..."
	./$(BUILD_DIR)/$(BINARY_NAME) keys add $(name) --keyring-backend test
	./$(BUILD_DIR)/$(BINARY_NAME) add-genesis-account $$(./$(BUILD_DIR)/$(BINARY_NAME) keys show $(name) -a --keyring-backend test) 1000000000ufluentum,1000000000stake --keyring-backend test

# Run testnet startup script (Linux/macOS)
.PHONY: testnet-script
testnet-script:
	@echo "Running testnet startup script..."
	@chmod +x scripts/start_testnet.sh
	./scripts/start_testnet.sh

# Run testnet startup script (Windows)
.PHONY: testnet-script-win
testnet-script-win:
	@echo "Running testnet startup script (Windows)..."
	powershell -ExecutionPolicy Bypass -File scripts/start_testnet.ps1

# Create a new account
.PHONY: keys-add
keys-add:
	@echo "Adding new account..."
	./$(BUILD_DIR)/$(BINARY_NAME) keys add $(name)

# List accounts
.PHONY: keys-list
keys-list:
	@echo "Listing accounts..."
	./$(BUILD_DIR)/$(BINARY_NAME) keys list

# Show account address
.PHONY: keys-show
keys-show:
	@echo "Showing account address..."
	./$(BUILD_DIR)/$(BINARY_NAME) keys show $(name) -a

# Add genesis account
.PHONY: add-genesis-account
add-genesis-account:
	@echo "Adding genesis account..."
	./$(BUILD_DIR)/$(BINARY_NAME) add-genesis-account $(address) $(coins)

# Collect genesis transactions
.PHONY: collect-gentxs
collect-gentxs:
	@echo "Collecting genesis transactions..."
	./$(BUILD_DIR)/$(BINARY_NAME) collect-gentxs

# Validate genesis
.PHONY: validate-genesis
validate-genesis:
	@echo "Validating genesis..."
	./$(BUILD_DIR)/$(BINARY_NAME) validate-genesis

# Export app state
.PHONY: export
export:
	@echo "Exporting app state..."
	./$(BUILD_DIR)/$(BINARY_NAME) export

# Reset the node
.PHONY: reset
reset:
	@echo "Resetting node..."
	./$(BUILD_DIR)/$(BINARY_NAME) tendermint unsafe-reset-all

# Show node status
.PHONY: status
status:
	@echo "Showing node status..."
	./$(BUILD_DIR)/$(BINARY_NAME) status

# Query account balance
.PHONY: query-balance
query-balance:
	@echo "Querying account balance..."
	./$(BUILD_DIR)/$(BINARY_NAME) query bank balances $(address)

# Send tokens
.PHONY: tx-send
tx-send:
	@echo "Sending tokens..."
	./$(BUILD_DIR)/$(BINARY_NAME) tx bank send $(from) $(to) $(amount) --chain-id $(chain-id) -y

# Create Fluentum record
.PHONY: tx-create-fluentum
tx-create-fluentum:
	@echo "Creating Fluentum record..."
	./$(BUILD_DIR)/$(BINARY_NAME) tx fluentum create-fluentum $(index) $(title) $(body) --chain-id $(chain-id) -y

# List Fluentum records
.PHONY: query-list-fluentum
query-list-fluentum:
	@echo "Listing Fluentum records..."
	./$(BUILD_DIR)/$(BINARY_NAME) query fluentum list-fluentum

# Show Fluentum record
.PHONY: query-show-fluentum
query-show-fluentum:
	@echo "Showing Fluentum record..."
	./$(BUILD_DIR)/$(BINARY_NAME) query fluentum show-fluentum $(index)

# Query Fluentum parameters
.PHONY: query-fluentum-params
query-fluentum-params:
	@echo "Querying Fluentum parameters..."
	./$(BUILD_DIR)/$(BINARY_NAME) query fluentum params

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	$(DOCKER) build -t fluentum:latest .

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	$(DOCKER) run -p 26656:26656 -p 26657:26657 -p 1317:1317 fluentum:latest

# Docker stop
.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	$(DOCKER) stop $$($(DOCKER) ps -q --filter ancestor=fluentum:latest)

# Help
.PHONY: help
help:
	@echo "Fluentum Core - Hybrid Consensus Blockchain Platform"
	@echo ""
	@echo "Dependency Management:"
	@echo "  deps                     - Download and tidy dependencies (auto-run by build)"
	@echo "  deps-check               - Check dependencies without modifying files"
	@echo ""
	@echo "Build Targets:"
	@echo "  build                    - Build binary with automatic dependency management"
	@echo "  build-only               - Build binary (assumes deps are ready)"
	@echo "  install                  - Install binary with automatic dependency management"
	@echo "  install-only             - Install binary (assumes deps are ready)"
	@echo "  clean                    - Clean build artifacts"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  test                     - Run tests"
	@echo "  test-coverage            - Run tests with coverage"
	@echo "  lint                     - Lint the code"
	@echo "  fmt                      - Format the code"
	@echo "  proto-gen                - Generate protobuf code"
	@echo ""
	@echo "Node Management:"
	@echo "  init-node                - Initialize a new node"
	@echo "  start                    - Start the node"
	@echo "  start-dev                - Start the node in development mode"
	@echo "  reset                    - Reset the node"
	@echo "  status                   - Show node status"
	@echo ""
	@echo "Testnet Management:"
	@echo "  init-testnet             - Initialize testnet node"
	@echo "  start-testnet            - Start testnet node"
	@echo "  start-testnet-bg         - Start testnet node in background"
	@echo "  stop-testnet             - Stop testnet node"
	@echo "  testnet-logs             - Show testnet logs"
	@echo "  reset-testnet            - Reset testnet node"
	@echo "  testnet-genesis-account name=NAME - Create testnet genesis account"
	@echo "  testnet-script           - Run testnet startup script (Linux/macOS)"
	@echo "  testnet-script-win       - Run testnet startup script (Windows)"
	@echo ""
	@echo "Account & Key Management:"
	@echo "  keys-add name=NAME       - Add a new account"
	@echo "  keys-list                - List accounts"
	@echo "  keys-show name=NAME      - Show account address"
	@echo ""
	@echo "Genesis & Chain Management:"
	@echo "  add-genesis-account address=ADDR coins=COINS - Add genesis account"
	@echo "  collect-gentxs           - Collect genesis transactions"
	@echo "  validate-genesis         - Validate genesis"
	@echo "  export                   - Export app state"
	@echo ""
	@echo "Query & Transaction Commands:"
	@echo "  query-balance address=ADDR - Query account balance"
	@echo "  tx-send from=FROM to=TO amount=AMT chain-id=CHAIN - Send tokens"
	@echo "  tx-create-fluentum index=IDX title=TITLE body=BODY chain-id=CHAIN - Create Fluentum record"
	@echo "  query-list-fluentum      - List Fluentum records"
	@echo "  query-show-fluentum index=IDX - Show Fluentum record"
	@echo "  query-fluentum-params    - Query Fluentum parameters"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build             - Build Docker image"
	@echo "  docker-run               - Run Docker container"
	@echo "  docker-stop              - Stop Docker container"
	@echo ""
	@echo "Development Notes:"
	@echo "  - 'build' automatically runs 'deps' to ensure dependencies are ready"
	@echo "  - Use 'build-only' in CI/CD when dependencies are pre-managed"
	@echo "  - Run 'deps' manually if you need to update dependencies"
	@echo "  - Use 'deps-check' to verify dependencies without changes"
