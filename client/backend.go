package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	xdr "github.com/davecgh/go-xdr/xdr2"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/ed25519"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/smrepl/common"
	"github.com/spacemeshos/smrepl/log"
	"github.com/spacemeshos/smrepl/smWallet"
	"golang.org/x/crypto/ssh/terminal"
)

// WalletBackend wallet holder
type WalletBackend struct {
	*gRPCClient      // Embedded interface
	workingDirectory string
	wallet           *smWallet.Wallet
	open             bool
}

func (w *WalletBackend) IsOpen() bool {
	return w.wallet != nil
}

func friendlyTime(nastyString string) string {
	t, err := time.Parse("2006-01-02T15-04-05.000Z", nastyString)
	if err != nil {
		return nastyString
	}
	return t.Format("Jan 02 2006 03:04 PM")
}

func (w *WalletBackend) PrintWalletMnemonic() {
	mnemonic, err := w.wallet.GetMnemonic()
	if err != nil {
		log.Error("error reading mnemonic", err)
		return
	}

	fmt.Println("Mnemonic:", mnemonic)
}

func (w *WalletBackend) WalletInfo() {
	fmt.Println("Name:", w.wallet.Meta.DisplayName)
	fmt.Println("Created:", friendlyTime(w.wallet.Meta.Created))
	fmt.Println("File Path:", w.wallet.WalletPath())

	addressesCount, err := w.wallet.GetNumberOfAccounts()
	if err != nil {
		log.Error("error reading addresses count", err)
		return
	}

	fmt.Println("Addresses created:", addressesCount)

}

func getString(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin)) // no history
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

func getClearString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)
	return strings.TrimSpace(text)
}

func getPassword() (string, error) {
	return getString("Enter wallet file password: ")
}

// OpenConnection opens a connection but not the wallet
func OpenConnection(grpcServer string, secureConnection bool, wd string) (wbx *WalletBackend, err error) {
	wbe := WalletBackend{workingDirectory: wd}
	wbe.gRPCClient = newGRPCClient(grpcServer, secureConnection)
	if err = wbe.gRPCClient.Connect(); err != nil {
		// failed to connect to grpc server
		log.Error("failed to connect to the grpc server: %s", err)
		return
	}
	return &wbe, nil
}

func accounts(num int) string {
	if num == 1 {
		return "1 account"
	}
	return fmt.Sprintf("%d accounts", num)
}

// OpenWallet opens a wallet from file
func (w *WalletBackend) OpenWallet() bool {
	fmt.Println("Press the TAB key to select a wallet file")
	walletToOpen := w.getWallet()
	wallet, err := smWallet.LoadWallet(walletToOpen)
	if err != nil {
		fmt.Println("Error loading wallet from file", err)
		return false
	}
	w.wallet = wallet
	password, err := getPassword()
	if err != nil {
		return false
	}
	fmt.Println("\nOpening wallet...")
	if err = w.wallet.Unlock(password); err != nil {
		fmt.Println("Wrong wallet password")
		return false
	}
	ne, err := w.wallet.GetNumberOfAccounts()
	if err != nil {
		fmt.Println("Invalid wallet file")
		return false
	}
	fmt.Println("Wallet ", w.wallet.Meta.DisplayName, "successfully opened with", accounts(ne))
	w.open = true
	return true
}

// OpenWalletBackend opens an existing wallet
func OpenWalletBackend(wallet string, grpcServer string, secureConnection bool) (wbx *WalletBackend, err error) {
	var wbe WalletBackend
	wbx = nil
	if wbe.wallet, err = smWallet.LoadWallet(wallet); err != nil {
		return
	}
	password, err := getPassword()
	if err != nil {
		return
	}
	fmt.Println("\nOpening wallet...")
	if err = wbe.wallet.Unlock(password); err != nil {
		return
	}
	ne, err := wbe.wallet.GetNumberOfAccounts()
	if err != nil {
		return nil, err
	}
	fmt.Println("Wallet", wbe.wallet.Meta.DisplayName, "successfully opened with", accounts(ne))
	wbe.open = true
	return &wbe, nil
}

func (w *WalletBackend) NewWallet() bool {
	walletName := getClearString("Wallet display name: ")
	fmt.Println()
	password, err := getPassword()
	fmt.Println()
	if err != nil {
		return false
	}
	password2, err := getString("Repeat password: ")
	fmt.Println()
	if err != nil {
		return false
	}
	if password != password2 {
		fmt.Println("Passwords do not match")
		return false
	}

	mnemonicString := getClearString("Use existing mnemonic (optional): ")
	fmt.Println()
	if len(mnemonicString) > 0 {
		w.wallet, err = smWallet.NewWalletWithMnemonic(walletName, password, mnemonicString)
		if err != nil {
			fmt.Println(err)
			return false
		}
	} else {
		w.wallet, err = smWallet.NewWallet(walletName, password)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}

	err = w.wallet.SaveWalletAs(w.workingDirectory + "/my_wallet")
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("Wallet successfully created")
	w.open = true
	return true
}

func (w *WalletBackend) CloseWallet() {
	w.wallet = nil
}

// CurrentAccount - get the latest account into cli-wallet format
func (w *WalletBackend) CurrentAccount() (*common.LocalAccount, error) {

	ca, err := w.wallet.CurrentAccount()
	if err != nil {
		return nil, err
	}
	pk, err := ca.PrivateKey()
	if err != nil {
		return nil, err
	}
	return &common.LocalAccount{Name: ca.DisplayName, PrivKey: pk, PubKey: smWallet.PublicKey(pk)}, nil
}

func (w *WalletBackend) CreateAccount(displayName string) (la *common.LocalAccount, err error) {
	pos, err := w.wallet.GenerateNewPair(displayName)
	if err != nil {
		return nil, err
	}
	if err = w.wallet.SetCurrent(pos); err != nil {
		return
	}
	return w.CurrentAccount()
}

func (w *WalletBackend) SetCurrentAccount(accountNumber int) error {
	return w.wallet.SetCurrent(accountNumber)
}

func interfaceToBytes(i interface{}) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &i); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (w *WalletBackend) StoreAccounts() error {
	return w.wallet.SaveWallet()
}

// Transfer creates a sign coin transaction and submits it
func (w *WalletBackend) Transfer(recipient gosmtypes.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (*pb.TransactionState, error) {
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

func (w *WalletBackend) GetAccount(accountName string) (*common.LocalAccount, error) {
	numberOfAccounts, err := w.wallet.GetNumberOfAccounts()
	if err != nil {
		log.Error("failed to retrieve number of accounts", err)
		return nil, err
	}
	for j := 0; j < numberOfAccounts; j++ {
		dn, err := w.wallet.GetAccountDisplayName(j)
		if err != nil {
			log.Error("failed to retrieve display names", err)
			return nil, err
		}
		if dn == accountName {
			pk, err := w.wallet.GetPrivateKey(j)
			if err != nil {
				log.Error("failed to retrieve private key", err)
				return nil, err
			}
			return &common.LocalAccount{Name: accountName, PrivKey: pk, PubKey: smWallet.PublicKey(pk)}, nil
		}
	}
	err = errors.New("failed to find :" + accountName)
	log.Error(err.Error())
	return nil, err
}

func (w *WalletBackend) ListAccounts() (res []string, err error) {
	numberOfAccounts, err := w.wallet.GetNumberOfAccounts()
	if err != nil {
		log.Error("failed to retrieve number of accounts", err)
		return []string{}, err
	}
	for j := 0; j < numberOfAccounts; j++ {
		dn, err := w.wallet.GetAccountDisplayName(j)
		if err != nil {
			log.Error("failed to retrieve display names", err)
			return []string{}, err
		}
		res = append(res, dn)
	}

	return res, nil
}
