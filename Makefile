PACKAGES=$(shell go list ./...)
OUTPUT?=build/fluentum

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

all: check build test install
.PHONY: all

include tests.mk

###############################################################################
###                                Build Fluentum                          ###
###############################################################################

build:
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o $(OUTPUT) ./cmd/fluentum/
.PHONY: build

install:
	CGO_ENABLED=$(CGO_ENABLED) go install $(BUILD_FLAGS) -tags $(BUILD_TAGS) ./cmd/fluentum
.PHONY: install


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

build_abci:
	@go build -mod=readonly -i ./abci/cmd/...
.PHONY: build_abci

install_abci:
	@go install -mod=readonly ./abci/cmd/...
.PHONY: install_abci

###############################################################################
###                              Distribution                               ###
###############################################################################

# dist builds binaries for all platforms and packages them for distribution
# TODO add abci to these scripts
dist:
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
BINARY_NAME := fluentum
MAIN_PACKAGE := ./cmd/fluentum
LDFLAGS := -X github.com/kellyadamtan/tendermint/version.Version=$(VERSION)

# Default target
.PHONY: all
all: clean build

# Build the binary
.PHONY: build
build:
	@echo "Building Fluentum Core $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "$(LDFLAGS)" $(MAIN_PACKAGE)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

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

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Generate protobuf files
.PHONY: proto
proto:
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/**/*.proto

# Help target
.PHONY: help
help:
	@echo "Fluentum Core Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build the binary"
	@echo "  make clean        Clean build artifacts"
	@echo "  make test         Run tests"
	@echo "  make lint         Run linter"
	@echo "  make deps         Install dependencies"
	@echo "  make proto        Generate protobuf files"
	@echo "  make help         Show this help message"

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
	@rm -rf $(BUILD_DIR)
	@rm -rf .fluentum

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
	@echo "Available targets:"
	@echo "  build                    - Build the binary"
	@echo "  install                  - Install the binary"
	@echo "  clean                    - Clean build artifacts"
	@echo "  test                     - Run tests"
	@echo "  test-coverage            - Run tests with coverage"
	@echo "  lint                     - Lint the code"
	@echo "  fmt                      - Format the code"
	@echo "  proto-gen                - Generate protobuf code"
	@echo "  init-node                - Initialize a new node"
	@echo "  start                    - Start the node"
	@echo "  start-dev                - Start the node in development mode"
	@echo "  keys-add name=NAME       - Add a new account"
	@echo "  keys-list                - List accounts"
	@echo "  keys-show name=NAME      - Show account address"
	@echo "  add-genesis-account address=ADDR coins=COINS - Add genesis account"
	@echo "  collect-gentxs           - Collect genesis transactions"
	@echo "  validate-genesis         - Validate genesis"
	@echo "  export                   - Export app state"
	@echo "  reset                    - Reset the node"
	@echo "  status                   - Show node status"
	@echo "  query-balance address=ADDR - Query account balance"
	@echo "  tx-send from=FROM to=TO amount=AMT chain-id=CHAIN - Send tokens"
	@echo "  tx-create-fluentum index=IDX title=TITLE body=BODY chain-id=CHAIN - Create Fluentum record"
	@echo "  query-list-fluentum      - List Fluentum records"
	@echo "  query-show-fluentum index=IDX - Show Fluentum record"
	@echo "  query-fluentum-params    - Query Fluentum parameters"
	@echo "  docker-build             - Build Docker image"
	@echo "  docker-run               - Run Docker container"
	@echo "  docker-stop              - Stop Docker container"
	@echo "  help                     - Show this help message"
