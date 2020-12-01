package main

import (
	"flag"
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

	grpcServerPort := uint(client.DefaultGRPCPort)
	grpcServer := client.DefaultGRPCServer

	flag.StringVar(&grpcServer, "grpc-server", grpcServer, "The Spacemesh api grpc server")
	flag.UintVar(&grpcServerPort, "grpc-port", grpcServerPort, "The Spacemesh api grpc server port")

	flag.Parse()

	_, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return
	}
	be, err := client.NewWalletBackend(dataDir, grpcServer, grpcServerPort)
	if err != nil {
		return
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
