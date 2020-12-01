package client

import (
	"bytes"
	xdr "github.com/davecgh/go-xdr/xdr2"
	"github.com/spacemeshos/CLIWallet/common"
	"github.com/spacemeshos/CLIWallet/log"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/ed25519"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"path"
)

const accountsFileName = "accounts.json"

type walletBackend struct {
	*gRPCClient // Embedded interface
	common.Store
	accountsFilePath string
	currentAccount   *common.LocalAccount
}

func NewWalletBackend(dataDir string, grpcServer string, grpcPort uint) (*walletBackend, error) {
	accountsFilePath := path.Join(dataDir, accountsFileName)
	acc, err := common.LoadAccounts(accountsFilePath)
	if err != nil {
		log.Error("cannot load account from file %s: %s", accountsFilePath, err)
		acc = &common.Store{}
	}

	grpcClient := newGRPCClient(grpcServer, grpcPort)
	err = grpcClient.Connect()
	if err != nil {
		// failed to connect to grpc server
		return nil, err
	}

	return &walletBackend{grpcClient, *acc, accountsFilePath, nil}, nil
}

func (w *walletBackend) CurrentAccount() *common.LocalAccount {
	return w.currentAccount
}

func (w *walletBackend) SetCurrentAccount(a *common.LocalAccount) {
	w.currentAccount = a
}

func interfaceToBytes(i interface{}) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &i); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (w *walletBackend) StoreAccounts() error {
	return common.StoreAccounts(w.accountsFilePath, &w.Store)
}

// Transfer creates a sign coin transaction and submits it
func (w *walletBackend) Transfer(recipient gosmtypes.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (*pb.TransactionState, error) {
	tx := common.SerializableSignedTransaction{}
	tx.AccountNonce = nonce
	tx.Amount = amount
	tx.Recipient = recipient
	tx.GasLimit = gasLimit
	tx.Price = gasPrice

	buf, _ := interfaceToBytes(&tx.InnerSerializableSignedTransaction)
	copy(tx.Signature[:], ed25519.Sign2(key, buf))
	b, err := interfaceToBytes(&tx)
	if err != nil {
		return nil, err
	}
	return w.SubmitCoinTransaction(b)
}
