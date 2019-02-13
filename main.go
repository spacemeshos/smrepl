//go-spacemesh is a golang implementation of the Spacemesh node.
//See - https://spacemesh.io
package main

import (
	"github.com/CLIWallet/client"
	"github.com/CLIWallet/repl"
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
	_, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return
	}
	be, err := client.NewWalletBE()
	if err != nil {
		return
	}
	repl.Start(be)
}
