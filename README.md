# Fluentum Core

[![Go Version](https://img.shields.io/badge/Go-1.24.4+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Commits](https://img.shields.io/badge/Commits-8,761+-orange.svg)](https://github.com/fluentum-chain/fluentum/commits/main)
[![Code Size](https://img.shields.io/badge/Size-~150MB-lightgrey.svg)](https://github.com/fluentum-chain/fluentum)
[![CometBFT](https://img.shields.io/badge/CometBFT-v0.38.6-blue.svg)](https://cometbft.com/)
[![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-v0.50.6-green.svg)](https://cosmos.network/)

> **Next-Generation Hybrid Blockchain Platform** - High-performance, quantum-resistant, and privacy-enabled blockchain with cross-chain interoperability.

## 🚀 Overview

Fluentum Core is a production-ready blockchain platform that combines **Delegated Proof of Stake (DPoS)** with **Zero-Knowledge Rollups (ZK-Rollups)** for unprecedented performance and security. Built on **CometBFT v0.38.6** consensus with quantum-resistant cryptography and cross-chain capabilities.

### Key Differentiators

1. **🔄 Hybrid Consensus**: DPoS + ZK-Rollups for scalability and security
2. **🔐 Quantum-Resistant**: Post-quantum cryptography (Dilithium signatures)
3. **🌐 Cross-Chain**: Native interoperability with EVM and SVM chains
4. **⚡ High Performance**: Optimized for 10,000+ TPS
5. **🔒 Privacy**: Zero-knowledge proof integration
6. **🎯 Enterprise Ready**: Production-grade with comprehensive tooling
7. **🚀 ABCI++**: Full support for CometBFT's ABCI++ features

## 🔄 Migration to CometBFT

This project has been successfully migrated from Tendermint Core to **CometBFT v0.38.6** with **Cosmos SDK v0.50.6**. 

### Key Migration Changes

- ✅ **CometBFT v0.38.6**: Drop-in replacement for Tendermint v0.34+
- ✅ **Cosmos SDK v0.50.6**: Latest SDK with AutoCLI and PBTS
- ✅ **ABCI++ Support**: `PrepareProposal`, `ProcessProposal`, `ExtendVote`, `VerifyVoteExtension`
- ✅ **Proposer-Based Timestamps (PBTS)**: Enhanced timestamp handling
- ✅ **Nop Mempool**: Application-managed transaction handling
- ✅ **Pebble Database**: High-performance storage backend
- ✅ **Environment Variables**: `TMHOME` → `CMTHOME`

### Quick Migration

```bash
# Automatic migration (Linux/macOS)
chmod +x scripts/migrate-config.sh
./scripts/migrate-config.sh

# Windows PowerShell
.\scripts\migrate-config.ps1

# Manual migration
go install github.com/cometbft/confix@latest
confix migrate --home ~/.cometbft --target-version v0.38.6
```

For detailed migration instructions, see [Migration Guide](#migration-guide) below.

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
| **CometBFT Version** | v0.38.6 |
| **Cosmos SDK Version** | v0.50.6 |

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
├── 📁 scripts/            # Migration and utility scripts
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

# Install dependencies
go mod tidy

# Build with CometBFT and PebbleDB
make build

# Initialize a new node
./build/fluentumd init --home ~/.cometbft

# Start the node
./build/fluentumd start --home ~/.cometbft
```

### ✅ Verify Installation

```bash
# Check version
./build/fluentumd version

# Initialize a new node
./build/fluentumd init --home ~/.cometbft

# Start the node
./build/fluentumd start --home ~/.cometbft

# Show all commands
./build/fluentumd --help
```

## 🔄 Migration Guide

### For Existing Users

If you're upgrading from a previous version with Tendermint Core:

#### 1. Automatic Migration (Recommended)

**Linux/macOS:**
```bash
chmod +x scripts/migrate-config.sh
./scripts/migrate-config.sh
```

**Windows:**
```powershell
.\scripts\migrate-config.ps1
```

#### 2. Manual Migration

```bash
# 1. Install confix
go install github.com/cometbft/confix@latest

# 2. Set environment variables
export CMTHOME="$HOME/.cometbft"
export TMHOME="$HOME/.tendermint"

# 3. Migrate configuration
confix migrate --home $CMTHOME --target-version v0.38.6

# 4. Update dependencies
go mod tidy

# 5. Rebuild
make build
```

#### 3. Environment Variables

Update your shell profile:

```bash
# Remove old TMHOME
unset TMHOME

# Add new CMTHOME
export CMTHOME="$HOME/.cometbft"
```

**Windows:**
```powershell
# Remove TMHOME
[Environment]::SetEnvironmentVariable("TMHOME", $null, "User")

# Add CMTHOME
[Environment]::SetEnvironmentVariable("CMTHOME", "$env:USERPROFILE\.cometbft", "User")
```

### Configuration Changes

Key configuration updates in `config/config.toml`:

```toml
# Database backend (now pebble)
db_backend = "pebble"

# Mempool (now nop for application-managed transactions)
[mempool]
version = "nop"
recheck = false
broadcast = false

# Consensus with PBTS
[consensus]
pbts_enable = true
signature_scheme = "ed25519"
timeout_commit = "5s"

# Quantum features
[quantum]
enabled = true
mode = "dilithium3"
```

## 🎯 Core Features

### 🔄 Hybrid Consensus Engine
- **DPoS**: Delegated Proof of Stake for fast finality
- **ZK-Rollups**: Zero-knowledge proofs for scalability
- **Hybrid Router**: Intelligent transaction routing
- **ABCI++**: Full CometBFT ABCI++ support

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
- **Pebble Database**: High-performance storage

### 🔒 Privacy Features
- **Zero-Knowledge Proofs**: Privacy-preserving transactions
- **zk-KYC**: Privacy-compliant identity verification
- **Confidential Transactions**: Optional transaction privacy

### 🚀 ABCI++ Features

The application implements all ABCI++ methods for enhanced functionality:

- **PrepareProposal**: Custom transaction selection and ordering
- **ProcessProposal**: Transaction validation during proposal processing
- **ExtendVote**: Vote extensions for side transactions
- **VerifyVoteExtension**: Validation of vote extensions from other validators

## 🏗️ Architecture

### Core Components

| Component | Description | Status |
|-----------|-------------|--------|
| **Consensus Engine** | CometBFT v0.38.6 + Hybrid DPoS | ✅ Production |
| **Quantum Crypto** | Dilithium signatures | ✅ Implemented |
| **Cross-Chain Bridge** | EVM/SVM interoperability | 🔄 Development |
| **Privacy Layer** | ZK-proof integration | 🔄 Development |
| **Smart Contracts** | Solidity contracts | ✅ Ready |
| **RPC Interface** | JSON-RPC & gRPC | ✅ Complete |
| **ABCI++** | Full ABCI++ support | ✅ Complete |

### Technology Stack

- **Consensus**: CometBFT v0.38.6 + Custom DPoS
- **Application Framework**: Cosmos SDK v0.50.6
- **Cryptography**: Dilithium, Ed25519, Secp256k1
- **Smart Contracts**: Solidity (EVM) + Rust (SVM)
- **Networking**: P2P with libp2p
- **Storage**: PebbleDB (recommended), LevelDB, RocksDB
- **API**: JSON-RPC, gRPC, WebSocket

## 📚 Documentation

- **[Migration Guide](#migration-guide)** - CometBFT migration instructions
- **[Ubuntu Installation Guide](INSTALL_UBUNTU.md)** - Detailed Ubuntu setup
- **[Configuration Guide](docs/configuration.md)** - Node configuration
- **[API Documentation](docs/api.md)** - RPC and gRPC APIs
- **[Architecture Specs](docs/introduction/architecture.md)** - Technical architecture
- **[Smart Contracts](contracts/)** - Solidity contract documentation
- **[CometBFT Docs](https://docs.cometbft.com/)** - CometBFT documentation

## 🧪 Development

### Build Commands

```bash
# Build for development (with PebbleDB - recommended)
make build

# Build with specific tags
make build-tags="pebble"

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
go mod tidy

# 3. Build
make build

# 4. Test
make test

# 5. Run locally
./build/fluentumd start --home ~/.cometbft
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- [GitHub Issues](https://github.com/fluentum-chain/fluentum/issues)
- [Discord Community](https://discord.gg/fluentum)
- [Documentation](https://docs.fluentum.com)

## 🔗 Links

- [Website](https://fluentum.com)
- [Explorer](https://explorer.fluentum.com)
- [API Documentation](https://api.fluentum.com)
- [CometBFT Documentation](https://docs.cometbft.com/)
- [Cosmos SDK Documentation](https://docs.cosmos.network/)

---

<div align="center">

**Fluentum Core** - *Trade Crypto Fluidly* 🚀

[Website](https://fluentum.tech) • [Documentation](https://docs.fluentum.tech) • [Community](https://t.me/fluentum)

</div>
