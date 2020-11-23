package client

import (
	"bytes"
	"fmt"
	xdr "github.com/davecgh/go-xdr/xdr2"
	"github.com/spacemeshos/CLIWallet/accounts"
	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/ed25519"
	go_sm_types "github.com/spacemeshos/go-spacemesh/common/types"

	"path"
)

const accountsFileName = "accounts.json"

type WalletBE struct {
	grpcClient *GRPCClient
	*HTTPRequester
	accounts.Store
	accountsFilePath string
	currentAccount   *accounts.LocalAccount
}

func NewWalletBE(serverHostPort, datadir string, grpcServer string, grpcPort uint) (*WalletBE, error) {
	accountsFilePath := path.Join(datadir, accountsFileName)
	acc, err := accounts.LoadAccounts(accountsFilePath)
	if err != nil {
		log.Error("cannot load account from file %s: %s", accountsFilePath, err)
		acc = &accounts.Store{}
	}


	url := fmt.Sprintf("http://%s/v1", serverHostPort)
	return &WalletBE{NewGRPCClient(grpcServer, grpcPort), NewHTTPRequester(url), *acc, accountsFilePath, nil}, nil
}

func (w *WalletBE) CurrentAccount() *accounts.LocalAccount {
	return w.currentAccount
}

func (w *WalletBE) SetCurrentAccount(a *accounts.LocalAccount) {
	w.currentAccount = a
}

func InterfaceToBytes(i interface{}) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &i); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (w *WalletBE) StoreAccounts() error {
	return accounts.StoreAccounts(w.accountsFilePath, &w.Store)
}

func (w *WalletBE) Transfer(recipient go_sm_types.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (string, error) {
	tx := SerializableSignedTransaction{}
	tx.AccountNonce = nonce
	tx.Amount = amount
	tx.Recipient = recipient
	tx.GasLimit = gasLimit
	tx.Price = gasPrice

	buf, _ := InterfaceToBytes(&tx.InnerSerializableSignedTransaction)
	copy(tx.Signature[:], ed25519.Sign2(key, buf))
	b, err := InterfaceToBytes(&tx)
	if err != nil {
		return "", err
	}
	return w.HTTPRequester.Send(b)
}
