package client

import (
	"bytes"
	xdr "github.com/davecgh/go-xdr/xdr2"
	"github.com/spacemeshos/CLIWallet/local-types"
	"github.com/spacemeshos/CLIWallet/log"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/ed25519"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"path"
)

const accountsFileName = "accounts.json"

type WalletBackend struct {
	*GRPCClient // Embedded interface
	local_types.Store
	accountsFilePath string
	currentAccount   *local_types.LocalAccount
}

func NewWalletBackend(dataDir string, grpcServer string, grpcPort uint) (*WalletBackend, error) {
	accountsFilePath := path.Join(dataDir, accountsFileName)
	acc, err := local_types.LoadAccounts(accountsFilePath)
	if err != nil {
		log.Error("cannot load account from file %s: %s", accountsFilePath, err)
		acc = &local_types.Store{}
	}

	grpcClient := NewGRPCClient(grpcServer, grpcPort)
	err = grpcClient.Connect()
	if err != nil {
		// failed to connect to grpc server
		return nil, err
	}

	return &WalletBackend{grpcClient, *acc, accountsFilePath, nil}, nil
}

func (w *WalletBackend) CurrentAccount() *local_types.LocalAccount {
	return w.currentAccount
}

func (w *WalletBackend) SetCurrentAccount(a *local_types.LocalAccount) {
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
	return local_types.StoreAccounts(w.accountsFilePath, &w.Store)
}

// Transfer creates a sign coin transaction and submits it
func (w *WalletBackend) Transfer(recipient gosmtypes.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (*pb.TransactionState, error) {
	tx := local_types.SerializableSignedTransaction{}
	tx.AccountNonce = nonce
	tx.Amount = amount
	tx.Recipient = recipient
	tx.GasLimit = gasLimit
	tx.Price = gasPrice

	buf, _ := InterfaceToBytes(&tx.InnerSerializableSignedTransaction)
	copy(tx.Signature[:], ed25519.Sign2(key, buf))
	b, err := InterfaceToBytes(&tx)
	if err != nil {
		return nil, err
	}
	return w.SubmitCoinTransaction(b)
}
