//go-spacemesh is a golang implementation of the Spacemesh node.
//See - https://spacemesh.io
package main

import (
	"flag"
	"github.com/spacemeshos/CLIWallet/client"
	"github.com/spacemeshos/CLIWallet/repl"
	"github.com/spacemeshos/go-spacemesh/accounts"
	"os"
	"syscall"
)

type mockClient struct {
}

func (m mockClient) LocalAccount() *accounts.Account {
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

	flag.StringVar(&serverHostPort, "server", serverHostPort, "host:port of the Spacemesh node HTTP server")
	flag.StringVar(&datadir, "datadir", datadir, "The directory to store the wallet data within")
	flag.Parse()

	_, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return
	}
	be, err := client.NewWalletBE(serverHostPort, datadir)
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
