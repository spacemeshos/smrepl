package repl

import (
	"encoding/hex"
	"fmt"
	"github.com/spacemeshos/CLIWallet/accounts"
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
	CreateAccount(owner string) *accounts.Account
	LocalAccount() *accounts.Account
	SetLocalAccount(a *accounts.Account)
	AccountInfo(id string) (*accounts.AccountInfo, error)
	Transfer(recipient address.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) error
	ListAccounts() []string
	GetAccount(name string) (*accounts.Account, error)
	StoreAccounts() error
	NodeURL() string
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
		{"create account", "Create a new coin account", r.createAccount},
		{"account", "Display account info", r.account},
		{"transfer coins", "Transfer coins between any two accounts", r.transferCoins},
		{"switch account", "Switch to another account", r.chooseAccount},
		//{"unlock account", "Unlock account.", r.unlockAccount},
		//{"lock account", "Lock Account.", r.lockAccount},
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
	fmt.Println(accs)
	if len(accs) > 0 {
		r.chooseAccount()
	} else {
		r.createAccount()
	}
}

func (r *repl) chooseAccount() {
	accs := r.client.ListAccounts()
	accName := multipleChoice(accs)
	account, err := r.client.GetAccount(accName)
	if err != nil {
		panic("wtf")
	}
	fmt.Printf("%s Loaded account: %s pubkey: 0x%x \n", printPrefix, account.Name, account.PubKey)
	r.client.SetLocalAccount(account)
}

func (r *repl) createAccount() {
	//accountInfo := prompt.Input(prefix+accountInfoMsg,
	//emptyComplete,
	//prompt.OptionPrefixTextColor(prompt.LightGray))
	accName := ""
	accName = prompt.Input(prefix+createAccountMsg,
		emptyComplete,
		prompt.OptionPrefixTextColor(prompt.LightGray))

	ac := r.client.CreateAccount(accName)
	err := r.client.StoreAccounts()
	if err != nil {
		log.Error("%s", err)
	}
	fmt.Printf("%s Created account: %s pubkey: 0x%x \n", printPrefix, ac.Name, ac.PubKey)
	r.client.SetLocalAccount(ac)
}

func (r *repl) commandLineParams(idx int, input string) string {
	c := r.commands[idx]
	params := strings.Replace(input, c.text, "", -1)

	return strings.TrimSpace(params)
}

func (r *repl) account() {
	if r.client.LocalAccount() == nil {
		r.chooseAccount()
	}
	account := r.client.LocalAccount()
	info, err := r.client.AccountInfo(hex.EncodeToString(account.PubKey))
	if err != nil {
		log.Error("cannot find account: %s : %s", account.Name, err)
		return
	}
	fmt.Println(printPrefix, "Name: ", account.Name)
	fmt.Println(printPrefix, "Balance: ", info.Balance)
	fmt.Println(printPrefix, "Nonce: ", info.Nonce)
}

func (r *repl) transferCoins() {
	fmt.Println(printPrefix, initialTransferMsg)

	acct := r.client.LocalAccount()
	if acct == nil {
		accountCommand := r.commands[3]

		// executing account command to create a local account
		r.executor(accountCommand.text)
		return
	}

	accountID := hex.EncodeToString(acct.PubKey)

	destinationAccountID := multipleChoice(r.client.ListAccounts())
	dest, err := r.client.GetAccount(destinationAccountID)
	if err != nil {
		log.Error("unknown account")
		return
	}
	amountStr := inputNotBlank(amountToTransferMsg)

	acc, err := r.client.AccountInfo(accountID)
	if err != nil {
		log.Error("can't get client info")
		return
	}
	gas := uint64(1)
	if yesOrNoQuestion(useDefaultGasMsg) == "y" {
		gasStr := inputNotBlank(enterGasPrice)
		gas, err = strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			log.Error("invalid gas", err)
			return
		}
	}

	fmt.Println(printPrefix, "Transaction summary:")
	fmt.Println(printPrefix, "From:  ", address.BytesToAddress(acct.PubKey).String())
	fmt.Println(printPrefix, "To:    ", address.BytesToAddress(dest.PubKey).String())
	fmt.Println(printPrefix, "Amount:", amountStr)
	fmt.Println(printPrefix, "Gas:   ", gas)
	fmt.Println(printPrefix, "Nonce: ", acc.Nonce)

	nonce, err := strconv.ParseUint(acc.Nonce, 10, 32)
	amount, err := strconv.ParseUint(amountStr, 10, 32)

	if yesOrNoQuestion(confirmTransactionMsg) == "y" {
		err := r.client.Transfer(address.BytesToAddress(dest.PubKey), nonce, amount, gas, 100, acct.PrivKey)
		if err != nil {
			log.Info(err.Error())
			return
		}
	}
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
