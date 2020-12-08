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

## Using a public Spacemesh API server
You can use your wallet without running a full node by connecting it to a public Spacemesh api service for a Spacemesh network.
Use the `-grpc-server` and `-secure` flags connect to a remote Spacemesh api server. For example:

```bash
./cli_wallet_darwin_amd64 -server api-123.spacemesh.io:443 -secure
```

> Note that communications with the server will be secure using https/tls but the wallet doesn't currently verify the server identity.


## Using with a local full node

1. Join a Spacemesh network by running [go-spacemesh](https://github.com/spacemeshos/go-spacemesh/releases) or [Smapp](https://github.com/spacemeshos/smapp/releases) on your computer.
1. Build the wallet from this repository and run it. For example on OS X:

```bash
make build-mac
./cli_wallet_darwin_amd64
```

By default, the wallet attempts to connect to the api server provided by your locally running Spacemesh full node using the default node's grpc api port (localhost:9092).

When you run your full node directly in terminal, you can configure which api services will be available to your wallet by your node by changing entries int he api section of your node's config file:

```json
"api": {
    "grpc": "node, mesh, globalstate,transaction, smesher"
},
```
 





