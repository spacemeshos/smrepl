package client

import (
	"bytes"
	"github.com/spacemeshos/CLIWallet/accounts"
	"github.com/spacemeshos/CLIWallet/log"
	xdr "github.com/davecgh/go-xdr/xdr2"
	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/address"
)

const DataPath = "/tmp/"
const accountsPath = "accounts.json"

type WalletBE struct {
	*HTTPRequester
	accounts.Store
	localAccount *accounts.Account
}

func NewWalletBE(node string) (*WalletBE, error) {
	server := ServerAddress
	if node != "" {
		server = "http://" + node + "/v1"
	}
	acc, err := accounts.LoadAccounts(accountsPath)
	if err != nil {
		log.Error("cannot load account from file %s: %s", accountsPath, err)
		acc = &accounts.Store{}
	}
	return &WalletBE{NewHTTPRequester(server), *acc,nil}, nil
}

func (w *WalletBE) LocalAccount() *accounts.Account {
	return w.localAccount
}

func (w *WalletBE) SetLocalAccount(a *accounts.Account){
	w.localAccount = a
}

func InterfaceToBytes(i interface{}) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &i); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (w *WalletBE ) StoreAccounts() error{
	return accounts.StoreAccounts(accountsPath, &w.Store)
}

func (w *WalletBE) Transfer(recipient address.Address,nonce , amount, gasPrice, gasLimit uint64 ,key ed25519.PrivateKey) error {
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
		return err
	}
	return w.HTTPRequester.Send(b)
}
