package repl

import (
	"encoding/hex"
	"fmt"
	"github.com/spacemeshos/CLIWallet/accounts"
	"github.com/spacemeshos/CLIWallet/client"
	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/address"
	"os"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
)

const (
	prefix      = "$ "
	printPrefix = ">"
)

// TestMode variable used for check if unit test is running
var TestMode = false

type command struct {
	text        string
	description string
	fn          func()
}

type repl struct {
	commands []command
	client   Client
	input    string
}

// Client interface to REPL clients.
type Client interface {
	CreateAccount(alias string) *accounts.Account
	CurrentAccount() *accounts.Account
	SetCurrentAccount(a *accounts.Account)
	AccountInfo(address string) (*accounts.AccountInfo, error)
	NodeInfo() (*client.NodeInfo, error)
	Transfer(recipient address.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) error
	ListAccounts() []string
	GetAccount(name string) (*accounts.Account, error)
	StoreAccounts() error
	NodeURL() string
	Smesh(datadir string, space uint, coinbase string) error
	SetCoinbase(coinbase string) error

	//Unlock(passphrase string) error
	//IsAccountUnLock(id string) bool
	//Lock(passphrase string) error
	//SetVariables(params, flags []string) error
	//GetVariable(key string) string
	//Restart(params, flags []string) error
	//NeedRestartNode(params, flags []string) bool
	//Setup(allocation string) error
}

// Start starts REPL.
func Start(c Client) {
	if !TestMode {
		r := &repl{client: c}
		r.initializeCommands()

		runPrompt(r.executor, r.completer, r.firstTime, uint16(len(r.commands)))
	} else {
		// holds for unit test purposes
		hold := make(chan bool)
		<-hold
	}
}

func (r *repl) initializeCommands() {
	r.commands = []command{
		{"new", "Create a new account (key pair) and set as current", r.createAccount},
		{"set", "Set one of the previously created accounts as current", r.chooseAccount},
		{"info", "Display the current account info", r.accountInfo},
		{"net", "Display the node status", r.nodeInfo},
		{"tx", "Transfer coins from current account to another account", r.transferCoins},
		{"sign", "Sign a text message with the current account private key", r.sign},
		{"coinbase", "Set current account as coinbase account in the node", r.coinbase},
		{"smesh", "Start smeshing", r.smesh},

		//{"unlock accountInfo", "Unlock accountInfo.", r.unlockAccount},
		//{"lock accountInfo", "Lock Account.", r.lockAccount},
		//{"setup", "Setup POST.", r.setup},
		//{"restart node", "Restart node.", r.restartNode},
		//{"set", "change CLI flag or param. E.g. set param a=5 flag c=5 or E.g. set param a=5", r.setCLIFlagOrParam},
		//{"echo", "Echo runtime variable.", r.echoVariable},
	}
}

func (r *repl) executor(text string) {

	s := strings.TrimSpace(text)
	if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	for _, c := range r.commands {
		if len(text) >= len(c.text) && text[:len(c.text)] == c.text {
			r.input = text
			//log.Debug(userExecutingCommandMsg, c.text)
			c.fn()
			return
		}
	}

	fmt.Println(printPrefix, "invalid command.")
}

func (r *repl) completer(in prompt.Document) []prompt.Suggest {
	suggets := make([]prompt.Suggest, 0)
	for _, command := range r.commands {
		s := prompt.Suggest{
			Text:        command.text,
			Description: command.description,
		}

		suggets = append(suggets, s)
	}

	return prompt.FilterHasPrefix(suggets, in.GetWordBeforeCursor(), true)
}

func (r *repl) firstTime() {
	fmt.Println(printPrefix, splash)
	fmt.Println("Welcome to Spacemesh. Connected to node at ", r.client.NodeURL())
	accs := r.client.ListAccounts()
	if len(accs) > 0 {
		r.chooseAccount()
	} else {
		r.createAccount()
	}
}

func (r *repl) chooseAccount() {
	accs := r.client.ListAccounts()
	fmt.Println(printPrefix, "Choose an account to load:")
	accName := multipleChoice(accs)
	account, err := r.client.GetAccount(accName)
	if err != nil {
		panic("wtf")
	}
	fmt.Printf("%s Loaded account alias: `%s`, address: %s \n", printPrefix, account.Name, accounts.StringAddress(account.Address()))

	r.client.SetCurrentAccount(account)
}

func (r *repl) createAccount() {
	alias := inputNotBlank(createAccountMsg)

	ac := r.client.CreateAccount(alias)
	err := r.client.StoreAccounts()
	if err != nil {
		log.Error("failed to create account: %v", err)
		return
	}

	fmt.Printf("%s Created account alias: `%s`, address: %s \n", printPrefix, ac.Name, accounts.StringAddress(ac.Address()))
	r.client.SetCurrentAccount(ac)
}

func (r *repl) commandLineParams(idx int, input string) string {
	c := r.commands[idx]
	params := strings.Replace(input, c.text, "", -1)

	return strings.TrimSpace(params)
}

func (r *repl) accountInfo() {
	acct := r.client.CurrentAccount()
	if acct == nil {
		r.chooseAccount()
	}

	address := address.BytesToAddress(acct.PubKey)

	info, err := r.client.AccountInfo(hex.EncodeToString(address.Bytes()))
	if err != nil {
		log.Error("failed to get account info: %v", err)
		info = &accounts.AccountInfo{}
	}

	fmt.Println(printPrefix, "Local alias: ", acct.Name)
	fmt.Println(printPrefix, "Address: ", accounts.StringAddress(address))
	fmt.Println(printPrefix, "Balance: ", info.Balance)
	fmt.Println(printPrefix, "Nonce: ", info.Nonce)
	fmt.Println(printPrefix, fmt.Sprintf("Public key: 0x%s", hex.EncodeToString(acct.PubKey)))
	fmt.Println(printPrefix, fmt.Sprintf("Private key: 0x%s", hex.EncodeToString(acct.PrivKey)))
}

func (r *repl) nodeInfo() {
	acct := r.client.CurrentAccount()
	if acct == nil {
		r.chooseAccount()
	}

	info, err := r.client.NodeInfo()
	if err != nil {
		log.Error("failed to get node info: %v", err)
		return
	}

	fmt.Println(printPrefix, "Synced:", info.Synced)
	fmt.Println(printPrefix, "SyncedLayer:", info.SyncedLayer)
	fmt.Println(printPrefix, "CurrentLayer:", info.CurrentLayer)
	fmt.Println(printPrefix, "VerifiedLayer:", info.VerifiedLayer)
	fmt.Println(printPrefix, "Peers:", info.Peers)
	fmt.Println(printPrefix, "MinPeers:", info.MinPeers)
	fmt.Println(printPrefix, "MaxPeers:", info.MaxPeers)
	fmt.Println(printPrefix, "Smeshing datadir:", info.SmeshingDatadir)
	fmt.Println(printPrefix, "Smeshing status:", info.SmeshingStatus)
	fmt.Println(printPrefix, "Smeshing coinbase:", info.SmeshingCoinbase)
	fmt.Println(printPrefix, "Smeshing remainingBytes:", info.SmeshingRemainingBytes)
}

func (r *repl) transferCoins() {
	fmt.Println(printPrefix, initialTransferMsg)
	acct := r.client.CurrentAccount()
	if acct == nil {
		r.chooseAccount()
	}

	srcAddress := address.BytesToAddress(acct.PubKey)
	info, err := r.client.AccountInfo(hex.EncodeToString(srcAddress.Bytes()))
	if err != nil {
		log.Error("failed to get account info: %v", err)
		return
	}

	destAddressStr := inputNotBlank(destAddressMsg)
	destAddress := address.HexToAddress(destAddressStr)

	amountStr := inputNotBlank(amountToTransferMsg)

	gas := uint64(1)
	if yesOrNoQuestion(useDefaultGasMsg) == "n" {
		gasStr := inputNotBlank(enterGasPrice)
		gas, err = strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			log.Error("invalid gas", err)
			return
		}
	}

	fmt.Println(printPrefix, "Transaction summary:")
	fmt.Println(printPrefix, "From:  ", srcAddress.String())
	fmt.Println(printPrefix, "To:    ", destAddress.String())
	fmt.Println(printPrefix, "Amount:", amountStr)
	fmt.Println(printPrefix, "Gas:   ", gas)
	fmt.Println(printPrefix, "Nonce: ", info.Nonce)

	nonce, err := strconv.ParseUint(info.Nonce, 10, 32)
	amount, err := strconv.ParseUint(amountStr, 10, 32)

	if yesOrNoQuestion(confirmTransactionMsg) == "y" {
		err := r.client.Transfer(destAddress, nonce, amount, gas, 100, acct.PrivKey)
		if err != nil {
			log.Info(err.Error())
			return
		}
	}
}

func (r *repl) smesh() {
	acct := r.client.CurrentAccount()
	if acct == nil {
		r.chooseAccount()
	}

	datadir := inputNotBlank(smeshingDatadirMsg)

	spaceStr := inputNotBlank(smeshingSpaceAllocationMsg)
	space, err := strconv.ParseUint(spaceStr, 10, 32)
	if err != nil {
		log.Error("failed to parse: %v", err)
		return
	}

	if err := r.client.Smesh(datadir, uint(space)<<30, accounts.StringAddress(acct.Address())); err != nil {
		log.Error("failed to start smeshing: %v", err)
		return
	}
}

func (r *repl) coinbase() {
	acct := r.client.CurrentAccount()
	if acct == nil {
		r.chooseAccount()
	}

	if err := r.client.SetCoinbase(accounts.StringAddress(acct.Address())); err != nil {
		log.Error("failed to set coinbase: %v", err)
		return
	}
}

func (r *repl) sign() {
	acct := r.client.CurrentAccount()
	if acct == nil {
		r.chooseAccount()
	}

	msgStr := inputNotBlank(msgSignMsg)
	msg, err := hex.DecodeString(msgStr)
	if err != nil {
		log.Error("failed to decode msg hex string: %v", err)
		return
	}

	signature := ed25519.Sign2(acct.PrivKey, msg)

	fmt.Println(printPrefix, fmt.Sprintf("signature (in hex): %x", signature))
}

/*
func (r *repl) unlockAccount() {
	passphrase := r.commandLineParams(1, r.input)
	err := r.client.Unlock(passphrase)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	acctCmd := r.commands[3]
	r.executor(fmt.Sprintf("%s %s", acctCmd.text, passphrase))
}

func (r *repl) lockAccount() {
	passphrase := r.commandLineParams(2, r.input)
	err := r.client.Lock(passphrase)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	acctCmd := r.commands[3]
	r.executor(fmt.Sprintf("%s %s", acctCmd.text, passphrase))
}*/
