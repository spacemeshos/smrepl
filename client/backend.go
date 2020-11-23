package client

import (
	"bytes"
	xdr "github.com/davecgh/go-xdr/xdr2"
	"github.com/spacemeshos/CLIWallet/localtypes"
	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/ed25519"
	go_sm_types "github.com/spacemeshos/go-spacemesh/common/types"
	"path"
)

const accountsFileName = "accounts.json"

type WalletBackend struct {
	*GRPCClient // Embedded interface
	localtypes.Store
	accountsFilePath string
	currentAccount   *localtypes.LocalAccount
}

func NewWalletBackend(serverHostPort, datadir string, grpcServer string, grpcPort uint) (*WalletBackend, error) {
	accountsFilePath := path.Join(datadir, accountsFileName)
	acc, err := localtypes.LoadAccounts(accountsFilePath)
	if err != nil {
		log.Error("cannot load account from file %s: %s", accountsFilePath, err)
		acc = &localtypes.Store{}
	}

	grpcClient := NewGRPCClient(grpcServer, grpcPort)
	err = grpcClient.Connect()
	if err != nil {
		// failed to connect to grpc server
		return nil, err
	}

	return &WalletBackend{grpcClient, *acc, accountsFilePath, nil}, nil
}

func (w *WalletBackend) CurrentAccount() *localtypes.LocalAccount {
	return w.currentAccount
}

func (w *WalletBackend) SetCurrentAccount(a *localtypes.LocalAccount) {
	w.currentAccount = a
}

func InterfaceToBytes(i interface{}) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &i); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (w *WalletBackend) StoreAccounts() error {
	return localtypes.StoreAccounts(w.accountsFilePath, &w.Store)
}

func (w *WalletBackend) Transfer(recipient go_sm_types.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (string, error) {
	tx := localtypes.SerializableSignedTransaction{}
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
	return w.Send(b)
}
