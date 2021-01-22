package smWallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	xdr "github.com/davecgh/go-xdr/xdr2"
	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/common/util"
	"github.com/tyler-smith/go-bip39"
)

const (
	spaceSalt = "Spacemesh blockmesh"
	// ErrorNoFileName thrown if saving an unnamed wallet
	ErrorNoFileName = "No filename. Please use \"Save As\"."
	// ErrorWalletNotUnlocked thrown if attemt is made to access crypto data without unlocking
	ErrorWalletNotUnlocked = "Wallet has not been unlocked."
	// ErrorWalletDoesNotHaveThatAddress if attempting to access an account that has not been generated
	ErrorWalletDoesNotHaveThatAddress = "You are attempting to access an account that has not been generated."
	// ErrorWalletDoesNotHavePassword This wallet does not have a password.
	ErrorWalletDoesNotHavePassword = "Invalid State. This wallet does not have a password."
)

type account struct {
	DisplayName string `json:"displayName"`
	Created     string `json:"created"`
	Path        string `json:"path"`
	PublicKey   string `json:"publicKey"`
	SecretKey   string `json:"secretKey"`
}

func (a *account) Address() types.Address {
	return types.BytesToAddress(util.Hex2Bytes(a.PublicKey))
}

func (a *account) PrivateKey() (pub ed25519.PrivateKey, err error) {
	return hex.DecodeString(a.SecretKey)
}

func (w *Wallet) CurrentAccount() (*account, error) {
	if !w.unlocked {
		return nil, errors.New(ErrorWalletNotUnlocked)
	}
	return &w.Crypto.confidential.Accounts[w.Crypto.confidential.accountNumber], nil
}

type secretStuff struct {
	Mnemonic      string    `json:"mnemonic"`
	Accounts      []account `json:"accounts"`
	Contacts      []contact `json:"contacts"`
	accountNumber int
}

type contact struct {
	Nickname string `json:"nickname"`
	Address  string `json:"address"`
}

type walletMetadata struct {
	DisplayName string `json:"displayName"`
	Created     string `json:"created"`
	NetID       int    `json:"netId"`
	Meta        struct {
		Salt string `json:"salt"`
	} `json:"meta"`
}

type walletEncryptedData struct {
	Cipher       string `json:"cipher"`
	CipherText   string `json:"cipherText"`
	confidential secretStuff
}

// Wallet is the basic data structure.
type Wallet struct {
	keystore string
	password string
	unlocked bool
	Meta     walletMetadata      `json:"meta"`
	Crypto   walletEncryptedData `json:"crypto"`
}

// NewWallet returns a brand shiny new wallet with random seed and mnemonic phrase
func NewWallet(walletName, password string) (w *Wallet, err error) {
	wx := new(Wallet)
	wx.password = password
	wx.unlocked = true
	wx.Meta.Created = nowTimeString()
	wx.Meta.DisplayName = walletName
	wx.Meta.NetID = 0
	wx.Meta.Meta.Salt = spaceSalt
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}
	wx.Crypto.Cipher = "AES-128-CTR"
	wx.Crypto.confidential.Mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	wx.Crypto.confidential.accountNumber, err = wx.GenerateNewPair("Default")
	if err != nil {
		return nil, err
	}
	wx.Crypto.confidential.Contacts = []contact{}
	wx.reCrypt()
	return wx, nil
}

// LoadWallet returns a wallet object for an existing file copy of a wallet
func LoadWallet(keystore string) (w *Wallet, err error) {
	w = new(Wallet)
	var f *os.File
	f, err = os.Open(keystore)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(f).Decode(w)
	if err != nil {
		return nil, err
	}
	w.keystore = keystore
	err = w.verifyAccounts()
	w.Crypto.confidential.accountNumber = 0
	return
}

// SaveWalletAs saves a wallet to a file and records the filename internally
func (w *Wallet) SaveWalletAs(keystorePrefix string) (err error) {
	w.keystore = keystorePrefix + "_" + w.Meta.Created + ".json"
	return w.SaveWallet()
}

// SaveWallet saves a file only if it already has a filename
func (w *Wallet) SaveWallet() (err error) {
	if len(w.keystore) == 0 {
		return errors.New(ErrorNoFileName)
	}
	f, err := os.Create(w.keystore)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(w)
}

// Unlock a previously unlocked wallet
func (w *Wallet) Unlock(password string) (err error) {
	if w.unlocked {
		return nil
	}
	w.password = password
	ciphertext, err := hex.DecodeString(w.Crypto.CipherText)
	if err != nil {
		return
	}
	plaintextBytes, err := w.twoWayAES(ciphertext)
	if err != nil {
		return
	}
	err = json.Unmarshal(plaintextBytes, &w.Crypto.confidential)
	if err != nil {
		return err
	}
	w.unlocked = true
	return
}

func interfaceToBytes(i interface{}) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &i); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// SignedTransaction turns a transaction into a signed transaction :-)
func (w *Wallet) SignedTransaction(t *types.Transaction) ([]byte, error) {
	if !w.unlocked {
		return []byte{}, errors.New(ErrorWalletNotUnlocked)
	}
	acc, _ := w.CurrentAccount()
	key, _ := acc.PrivateKey()

	tx := struct {
		AccountNonce uint64
		Recipient    types.Address
		GasLimit     uint64
		Price        uint64
		Amount       uint64
	}{
		AccountNonce: t.AccountNonce,
		Amount:       t.Amount,
		Recipient:    t.Recipient,
		GasLimit:     t.GasLimit,
		Price:        t.Fee,
	}
	fmt.Println(tx.Recipient.Hex())
	buf, _ := interfaceToBytes(&tx)
	buf = append(buf, ed25519.Sign2(key, buf)...)
	return buf, nil
}

func (w *Wallet) WalletPath() string {
	return w.keystore
}
