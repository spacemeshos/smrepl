# Spacemesh CLIWallet

## Overview
A basic reference console Spacemesh Wallet designed to work together with a [go-spacemesh full node](https://github.com/spacemeshos/go-spacemesh).

The CLIWallet is designed for developers who want to hack on the Spacemesh platform. For non-devs we recommend using Smapp - the Spacemesh App. [Smapp](https://github.com/spacemeshos/smapp) is available for all major desktop platforms and includes a wallet.

## Functionality
The wallet is a Spacemesh API client and implements basic wallet features. You can create a new coin account, execute transactions, check account balance and view transactions./

> WARNING: CLIWallet is currently insecure as it saves private keys in cleartext on your local store. It is not yet a production-quality wallet. We plan to update the project to store private data securely in future releases. See [this issue](https://github.com/spacemeshos/CLIWallet/issues/16).

The CLI wallet implements this [mini spec](https://github.com/spacemeshos/product/blob/master/cli_wallet_spec.md). See below on how to use with any network.

## Building

### Build for your current platform with go:

```bash
go get && go build
```

### Build for all platforms:
```bash
make
```

### Build for a specific platforms:
```bash
make build-win
```

```bash
make build-mac
```

```bash
make build-linux
```

### With `docker`:
```
make dockerbuild-go
```
---

## Using with a local full node and an open testnet

1. Follow the instructions in the readme to build and run [go-spacemesh](https://github.com/spacemeshos/go-spacemesh) from source code.

2. Build the CLI Wallet from this repository and run it:

```bash
./cli_wallet
```

By default, the wallet will attempt to connect to a locally running Spacemesh full node using the default node's grpc api port.
You can configure the ports in go-sm and use CLIWallet `--grpc-server` and `--grpc-port` flags to override the defaults.
To use the full wallet features, enable all GRPC services in your node's config file. e.g:

```json
"api": {
        "grpc": "node, mesh, globalstate, transaction, smesher, debug"
    },
```

You can also use a full node managed by Smapp. Just start smapp to start a managed node and connect to it by running the CLIWallet from the command line.
You should be able to connect to the managed node without having to override the default CLIWallet settings.

## Using with a Local testnet
Please follow the instructions in our [local testnet guide](https://testnet.spacemesh.io/#/local)
