//go-spacemesh is a golang implementation of the Spacemesh node.
//See - https://spacemesh.io
package main

import (
	"flag"
	"github.com/spacemeshos/CLIWallet/localtypes"
	"github.com/spacemeshos/CLIWallet/client"
	"github.com/spacemeshos/CLIWallet/repl"
	"os"
	"syscall"
)

type mockClient struct {
}

func (m mockClient) LocalAccount() *localtypes.LocalAccount {
	return nil
}

func (m mockClient) AccountInfo(id string) {

}
func (m mockClient) Transfer(from, to, amount, passphrase string) error {
	return nil
}

func main() {
	serverHostPort := client.DefaultNodeHostPort
	datadir := Getwd()

	grpcServerPort := uint(client.DefaultGRPCPort)
	grpcServer := client.DefaultGRPCServer

	flag.StringVar(&serverHostPort, "server", serverHostPort, "host:port of the Spacemesh node HTTP server")
	flag.StringVar(&datadir, "datadir", datadir, "The directory to store the wallet data within")
	flag.StringVar(&grpcServer, "grpc-server", grpcServer, "The api 2.0 grpc server")
	flag.UintVar(&grpcServerPort, "grpc-port", grpcServerPort, "The api 2.0 grpc server port")

	flag.Parse()

	_, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return
	}
	be, err := client.NewWalletBackend(serverHostPort, datadir, grpcServer, grpcServerPort)
	if err != nil {
		return
	}
	repl.Start(be)
}

func Getwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}
