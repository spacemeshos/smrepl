package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spacemeshos/CLIWallet/client"
	"github.com/spacemeshos/CLIWallet/repl"
)

func main() {

	var (
		dataDir    string
		walletName string
		be         *client.WalletBackend
	)
	grpcServer := client.DefaultGRPCServer
	secureConnection := client.DefaultSecureConnection

	flag.StringVar(&grpcServer, "server", grpcServer, fmt.Sprintf("The Spacemesh api grpc server host and port. Defaults to %s", client.DefaultGRPCServer))
	flag.BoolVar(&secureConnection, "secure", secureConnection, "Connect securely to the server. Default is false")
	flag.StringVar(&dataDir, "wallet_directory", getwd(), "set default wallet directory")
	flag.StringVar(&walletName, "wallet", "", "set the name of wallet to open")

	flag.Parse()

	be, err := client.OpenConnection(grpcServer, secureConnection, dataDir)
	if err != nil {
		os.Exit(1)
	}
	if walletName != "" {
		walletPath := dataDir + "/" + walletName
		fmt.Println("opening ", walletPath)
		be, err = client.OpenWalletBackend(walletPath, grpcServer, secureConnection)
		if err != nil {
			fmt.Println("failed to open wallet : ", err)
			os.Exit(1)
		}
	}

	repl.Start(be)
}

func getwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}
