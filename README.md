# CLIWallet

## Overview
The CLI Wallet app currently only supports the [local testnet](https://github.com/spacemeshos/local-testnet/).

## Building
with `go`
```
go get && go build
```
with `docker`

`make dockerbuild-go`

---

## Using with a local full node and an open testnet
1. Build [go-spacemesh](https://github.com/spacemeshos/go-spacemesh) from source code.
2. Obtain the testnet's toml config file.
3. Start go-spacemesh with the following arguments:

```bash
./go-spacemesh --grpc-server --json-server --tcp-port [a_port] --config [tomlFileLocation] -d [nodeDataFilesPath]
```

For example:
```bash
./go-spacemesh --grpc-server --json-server --tcp-port 7153 --config tn1.toml -d ~/spacemesh_data
```

4. Build the CLI Wallet from this repository and run it:

```bash
./cli_wallet
```

---

## Using with a Local testnet
Please follow the instructions in our [local testnet guide](https://testnet.spacemesh.io/#/local)
