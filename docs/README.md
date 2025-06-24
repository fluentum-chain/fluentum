---
title: Tendermint Core Documentation
description: Tendermint Core is a blockchain application platform.
footer:
  newsletter: false
---

> **NOTICE: This documentation is for CometBFT v0.38+ and ABCI 2.0 (ABCI++).**
>
> - All code, APIs, and examples use the new ABCI 2.0 (ABCI++) interface.
> - Legacy ABCI 1.0 methods (`BeginBlock`, `DeliverTx`, `EndBlock`, etc.) are no longer supported. Use `FinalizeBlock` for all block-level processing in ABCI 2.0.
> - See the [main README](../README.md) and migration guide for details on upgrading.

# Tendermint

Welcome to the Tendermint Core documentation!

Tendermint Core is a blockchain application platform; it provides the equivalent
of a web-server, database, and supporting libraries for blockchain applications
written in any programming language. Like a web-server serving web applications,
Tendermint serves blockchain applications.

More formally, Tendermint Core performs Byzantine Fault Tolerant (BFT) State
Machine Replication (SMR) for arbitrary deterministic, finite state machines.
For more background, see [What is
Tendermint?](introduction/what-is-tendermint.md).

To get started quickly with an example application, see the [quick start
guide](introduction/quick-start.md).

To learn about application development on Tendermint, see the [Application
Blockchain
Interface](https://github.com/tendermint/tendermint/tree/v0.34.x/spec/abci).

For more details on using Tendermint, see the respective documentation for
[Tendermint Core](tendermint-core/), [benchmarking and monitoring](tools/), and
[network deployments](networks/).

To find out about the Tendermint ecosystem you can go
[here](https://github.com/tendermint/awesome#ecosystem). If you are a project
that is using Tendermint you are welcome to make a PR to add your project to the
list.

## Contribute

To contribute to the documentation, see [this
file](https://github.com/tendermint/tendermint/blob/main/docs/DOCS_README.md)
for details of the build process and considerations when making changes.
