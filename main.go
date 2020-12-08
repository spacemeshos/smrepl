package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/spacemeshos/CLIWallet/client"
	"github.com/spacemeshos/CLIWallet/common"
	"github.com/spacemeshos/CLIWallet/repl"
)

type mockClient struct {
}

func (m mockClient) LocalAccount() *common.LocalAccount {
	return nil
}

func (m mockClient) AccountInfo(id string) {

}
func (m mockClient) Transfer(from, to, amount, passphrase string) error {
	return nil
}

func main() {
	dataDir := getwd()
	grpcServer := client.DefaultGRPCServer
	secureConnection := client.DefaultSecureConnection

	flag.StringVar(&grpcServer, "server", grpcServer, fmt.Sprintf("The Spacemesh api grpc server host and port. Defaults to %s", client.DefaultGRPCServer))
	flag.BoolVar(&secureConnection, "secure", secureConnection, "Connect securely to the server. Default is false")

	flag.Parse()

	_, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		println(err)
		os.Exit(1)
	}
	be, err := client.NewWalletBackend(dataDir, grpcServer, secureConnection)
	if err != nil {
		println(err)
		os.Exit(1)
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
