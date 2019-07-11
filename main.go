//go-spacemesh is a golang implementation of the Spacemesh node.
//See - https://spacemesh.io
package main

import (
	"flag"
	"github.com/spacemeshos/CLIWallet/client"
	"github.com/spacemeshos/CLIWallet/repl"
	"github.com/spacemeshos/go-spacemesh/accounts"
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

func main() { // run the app
	wordPtr := flag.String("node", "", "node ip in port in format of IP:[PORT]")
	flag.Parse()
	_, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return
	}
	be, err := client.NewWalletBE(*wordPtr)
	if err != nil {
		return
	}
	repl.Start(be)
}
