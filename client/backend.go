package client

import (
	"fmt"
	"github.com/CLIWallet/accounts"
)

const DataPath = "/tmp/"

type WalletBE struct {
	*HTTPRequester
	localAccount *accounts.Account
}

func NewWalletBE() (*WalletBE, error) {
	err := accounts.LoadAllAccounts(DataPath)
	if err != nil {
		return nil, err
	}
	return &WalletBE{NewHTTPRequester(ServerAddress), accounts.LocalAccount}, nil
}

func (w *WalletBE) CreateAccount(passphrase string) error {
	acc, err := accounts.NewAccount(passphrase)
	if err != nil {
		return err
	}
	acc.Persist(DataPath)
	return nil
}

func (w *WalletBE) LocalAccount() *accounts.Account {
	return w.localAccount
}

func (w *WalletBE) Unlock(passphrase string) error {
	if w.localAccount == nil {
		return fmt.Errorf("no local account")
	}
	w.localAccount.UnlockAccount(passphrase)
	return nil
}

func (w *WalletBE) IsAccountUnLock(id string) bool {
	w.localAccount.IsAccountLocked()
	return false
}

func (w *WalletBE) Lock(passphrase string) error {
	if w.localAccount == nil {
		return fmt.Errorf("no local account")
	}
	w.localAccount.LockAccount(passphrase)
	return nil
}
