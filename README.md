# Fluentum Core

[![Go Version](https://img.shields.io/badge/Go-1.24.4+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Commits](https://img.shields.io/badge/Commits-8,761+-orange.svg)](https://github.com/fluentum-chain/fluentum/commits/main)
[![Code Size](https://img.shields.io/badge/Size-~150MB-lightgrey.svg)](https://github.com/fluentum-chain/fluentum)

> **Next-Generation Hybrid Blockchain Platform** - High-performance, quantum-resistant, and privacy-enabled blockchain with cross-chain interoperability.

## 🚀 Overview

Fluentum Core is a production-ready blockchain platform that combines **Delegated Proof of Stake (DPoS)** with **Zero-Knowledge Rollups (ZK-Rollups)** for unprecedented performance and security. Built on Tendermint consensus with quantum-resistant cryptography and cross-chain capabilities.

### Key Differentiators

1. **🔄 Hybrid Consensus**: DPoS + ZK-Rollups for scalability and security
2. **🔐 Quantum-Resistant**: Post-quantum cryptography (Dilithium signatures)
3. **🌐 Cross-Chain**: Native interoperability with EVM and SVM chains
4. **⚡ High Performance**: Optimized for 10,000+ TPS
5. **🔒 Privacy**: Zero-knowledge proof integration
6. **🎯 Enterprise Ready**: Production-grade with comprehensive tooling

## 📊 Project Statistics

| Metric | Value |
|--------|-------|
| **Repository Size** | ~150 MB (compressed) |
| **Source Code** | ~60 MB (uncompressed) |
| **Go Files** | 682 files (~5.4 MB) |
| **Total Files** | 1,205 files |
| **Directories** | 319 |
| **Git Commits** | 8,761+ |
| **Languages** | Go (82.2%), Solidity (2.6%), TeX (6.7%) |

### Architecture Components

```
fluentum/
├── 📁 consensus/          # Hybrid consensus (DPoS + ZK-Rollups)
├── 📁 crypto/             # Quantum-resistant cryptography
├── 📁 fluentum/           # Core Fluentum-specific modules
├── 📁 contracts/          # Smart contracts (Solidity)
├── 📁 circuits/           # Zero-knowledge circuits
├── 📁 cmd/fluentum/       # Main executable
├── 📁 docs/               # Comprehensive documentation
└── 📁 networks/           # Network configurations
```

## 🛠️ Quick Start

### Prerequisites
- **Go**: 1.24.4 or later
- **Git**: Latest version
- **System**: Ubuntu 20.04+ (recommended) or Windows/macOS

### 🚀 Automated Installation (Ubuntu)

```bash
# Clone the repository
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum

# Run automated installation
chmod +x install-ubuntu.sh
./install-ubuntu.sh
```

### 🔧 Manual Installation

```bash
# Clone the repository
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum

# Build and install with BadgerDB (recommended for Ubuntu)
CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make build
CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make install

# Alternative: Build with default settings
make build
make install
```

### ✅ Verify Installation

```bash
# Check version
fluentum version

# Initialize a new node
fluentum init

# Start the node
fluentum node

# Show all commands
fluentum --help
```

## 🎯 Core Features

### 🔄 Hybrid Consensus Engine
- **DPoS**: Delegated Proof of Stake for fast finality
- **ZK-Rollups**: Zero-knowledge proofs for scalability
- **Hybrid Router**: Intelligent transaction routing

### 🔐 Quantum-Resistant Security
- **Dilithium Signatures**: Post-quantum cryptography
- **Lattice-Based Crypto**: Future-proof security
- **Multi-Signature Support**: Enhanced security models

### 🌐 Cross-Chain Interoperability
- **EVM Compatibility**: Ethereum Virtual Machine support
- **SVM Support**: Solana Virtual Machine integration
- **Bridge Infrastructure**: Seamless asset transfers

### ⚡ Performance Optimizations
- **High Throughput**: 10,000+ TPS target
- **Fast Finality**: Sub-second block finality
- **Optimized Networking**: P2P optimization

### 🔒 Privacy Features
- **Zero-Knowledge Proofs**: Privacy-preserving transactions
- **zk-KYC**: Privacy-compliant identity verification
- **Confidential Transactions**: Optional transaction privacy

## 🏗️ Architecture

### Core Components

| Component | Description | Status |
|-----------|-------------|--------|
| **Consensus Engine** | Hybrid DPoS + ZK-Rollups | ✅ Production |
| **Quantum Crypto** | Dilithium signatures | ✅ Implemented |
| **Cross-Chain Bridge** | EVM/SVM interoperability | 🔄 Development |
| **Privacy Layer** | ZK-proof integration | 🔄 Development |
| **Smart Contracts** | Solidity contracts | ✅ Ready |
| **RPC Interface** | JSON-RPC & gRPC | ✅ Complete |

### Technology Stack

- **Consensus**: Tendermint Core + Custom DPoS
- **Cryptography**: Dilithium, Ed25519, Secp256k1
- **Smart Contracts**: Solidity (EVM) + Rust (SVM)
- **Networking**: P2P with libp2p
- **Storage**: LevelDB, RocksDB, BadgerDB
- **API**: JSON-RPC, gRPC, WebSocket

## 📚 Documentation

- **[Ubuntu Installation Guide](INSTALL_UBUNTU.md)** - Detailed Ubuntu setup
- **[Configuration Guide](docs/configuration.md)** - Node configuration
- **[API Documentation](docs/api.md)** - RPC and gRPC APIs
- **[Architecture Specs](docs/introduction/architecture.md)** - Technical architecture
- **[Smart Contracts](contracts/)** - Solidity contract documentation

## 🧪 Development

### Build Commands

```bash
# Build for development (with BadgerDB - recommended)
CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make build

# Build with default settings
make build

# Run tests
make test

# Format code
make format

# Lint code
make lint

# Generate protobuf
make proto-gen

# Build for specific platform
make build-linux
```

### Development Workflow

```bash
# 1. Clone and setup
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum

# 2. Install dependencies
go mod download

# 3. Build (with BadgerDB - no CGO dependencies)
CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make build

# 4. Test
make test

# 5. Run locally
./build/fluentum init
./build/fluentum node
```

## 🧩 Modular Feature System

Fluentum supports a **modular feature system** for advanced capabilities like quantum signing, state sync, and zk-rollup. Each feature is an independent Go module and can be enabled, disabled, or updated independently.

### Directory Structure

```
fluentum/
└── features/
    ├── quantum_signing/   # CRYSTALS-Dilithium quantum signatures
    ├── state_sync/        # Fast state synchronization
    └── zk_rollup/         # Zero-knowledge rollup
```

Each feature contains:
- `go.mod` — Go module definition
- `feature.go` — Feature implementation
- `build.sh` — Build/test script

### Build & Test Features

Build all features:
```bash
make features
```

Build a specific feature:
```bash
make feature FEATURE=quantum_signing
```

Test all features:
```bash
make test-features
```

### Runtime Configuration

Features are configured in `config/features.toml`:
```toml
[features.quantum_signing]
enabled = true
# CRYSTALS-Dilithium mode (1, 3, or 5)
dilithium_mode = 3
quantum_headers = true
max_latency_ms = 50

[features.state_sync]
enabled = false
fast_sync = true
chunk_size = 1000

[features.zk_rollup]
enabled = false
enable_proofs = true
batch_size = 100
```

- Enable/disable features by setting `enabled = true/false`.
- Tune feature-specific parameters as needed.

### Live Reloading & Version Compatibility
- Features are loaded and started at node startup.
- Hot reloading and version compatibility checks are supported.
- Feature updates can be distributed via Git submodules.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for:

- **Code of Conduct**: Community guidelines
- **Development Setup**: Local development environment
- **Pull Request Process**: How to submit changes
- **Testing Guidelines**: Code quality standards

### Development Areas

- 🔄 **Consensus**: Hybrid consensus optimization
- 🔐 **Cryptography**: Quantum-resistant implementations
- 🌐 **Interoperability**: Cross-chain bridge development
- 🔒 **Privacy**: ZK-proof integration
- 📊 **Performance**: Throughput optimization

## 📈 Roadmap

### Phase 1: Core Infrastructure ✅
- [x] Hybrid consensus implementation
- [x] Quantum-resistant cryptography
- [x] Basic cross-chain functionality
- [x] Smart contract support

### Phase 2: Advanced Features 🔄
- [ ] Enhanced privacy layer
- [ ] Advanced ZK-rollups
- [ ] Cross-chain bridges
- [ ] Governance system

### Phase 3: Ecosystem 🎯
- [ ] DeFi integrations
- [ ] NFT marketplace
- [ ] Enterprise solutions
- [ ] Mobile SDK

## 📄 License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

## 🆘 Support & Community

### Resources
- **🌐 Website**: [fluentum.tech](https://fluentum.tech)
- **📖 Documentation**: [docs.fluentum.tech](https://docs.fluentum.tech)
- **💬 Community**: [Telegram](https://t.me/fluentum)
- **🐛 Issues**: [GitHub Issues](https://github.com/fluentum-chain/fluentum/issues)

### Contact
- **Email**: support@fluentum.tech
- **Discord**: [Fluentum Community](https://discord.gg/fluentum)
- **Twitter**: [@FluentumChain](https://twitter.com/FluentumChain)

## Migration to CometBFT and Cosmos SDK v0.47+

### Config Migration (Confix)
- Install confix: `go get github.com/cosmos/confix@latest`
- To migrate or update your config file, run:
  ```sh
  confix merge --config config/config.toml --template config/config.template.toml --output config/config.toml
  ```
- Use confix for all future config merges/updates.

### Database Backend Update
- The default database backend is now `pebble` for CometBFT compatibility.
- If you have existing data, migrate it with:
  ```sh
  appd migrate --db-backend pebble
  ```
  (Replace `appd` with your binary name.)
- Update your `config.toml` to:
  ```toml
  db_backend = "pebble"
  ```

---

<div align="center">

**Fluentum Core** - *Trade Crypto Fluidly* 🚀

[Website](https://fluentum.tech) • [Documentation](https://docs.fluentum.tech) • [Community](https://t.me/fluentum)

</div>
