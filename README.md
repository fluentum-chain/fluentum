# Fluentum Core

Fluentum Mainnet aims to be a high-performance, interoperable, and privacy-enabled blockchain that differentiate from existing CEX mainnets by:

1. Supporting both EVM and SVM for broader developer adoption.
2. Offering additional privacy features.
3. Achieving higher throughput and faster finality.
4. Deep integration with the Fluentum ecosystem.

Fluentum Exchange is a cryptocurrency exchange combining CEX efficiency with DEX security, focusing on tokenized real-world assets and cross-chain interoperability.
and is positioned as a Next-Generation Super Exchange Ecosystem with: fluentum core engine, fluentum wallet, fluentum Blockchain mainnet, fluentum token $FLU as well as Zero Fees, Hybrid Liquidity, zk-KYC, AI Yield, Cross-Chain Gas, Compliance Oracle and Quantum Security.

Some of integrations supported by fluentum exchange:
- NFT Floor price Oracle
- RWA Tokenization Bridge
- On-Chain Analytics Pipeline
- Options Market Maker
- Staking Derivative Engine

Fluentum Exchange: 'Trade Crypto Fluidly'
www.fluentum.tech

## Quick Start

### Prerequisites
- Go 1.24.4 or later
- Git

### Installation

#### Ubuntu/Debian
```bash
# Clone the repository
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum

# Run automated installation
chmod +x install-ubuntu.sh
./install-ubuntu.sh
```

#### Manual Installation
```bash
# Clone the repository
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum

# Build and install
make build
make install
```

### Usage

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

## Documentation

- [Ubuntu Installation Guide](INSTALL_UBUNTU.md)
- [Configuration Guide](docs/configuration.md)
- [API Documentation](docs/api.md)

## Features

- **Hybrid Consensus**: Combines DPoS and ZK-Rollups
- **Quantum-Resistant**: Post-quantum cryptography support
- **Cross-Chain**: Interoperability with multiple blockchains
- **High Performance**: Optimized for high throughput
- **Privacy**: Zero-knowledge proof integration

## Development

```bash
# Build for development
make build

# Run tests
make test

# Format code
make format

# Lint code
make lint
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- GitHub Issues: https://github.com/fluentum-chain/fluentum/issues
- Documentation: https://docs.fluentum.tech
- Community: https://t.me/fluentum
