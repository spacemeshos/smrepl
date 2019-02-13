package repl

import (
	"fmt"
	"github.com/CLIWallet/accounts"
	"github.com/CLIWallet/log"
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
	CreateAccount(passphrase string) error
	LocalAccount() *accounts.Account
	Unlock(passphrase string) error
	IsAccountUnLock(id string) bool
	Lock(passphrase string) error
	AccountInfo(id string) (*accounts.AccountInfo, error)
	Transfer(from, to, amount, nonce, passphrase string) error
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
		{"create account", "Create account.", r.createAccount},
		{"unlock account", "Unlock account.", r.unlockAccount},
		{"lock account", "Lock Account.", r.lockAccount},
		{"account", "Shows basic local account info about the local account.", r.account},
		{"transfer coins", "Transfer coins between 2 accounts.", r.transferCoins},
		//{"setup", "Setup POST.", r.setup},
		//{"restart node", "Restart node.", r.restartNode},
		//{"set", "change CLI flag or param. E.g. set param a=5 flag c=5 or E.g. set param a=5", r.setCLIFlagOrParam},
		//{"echo", "Echo runtime variable.", r.echoVariable},
	}
}

func (r *repl) executor(text string) {
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
	createNewAccount := yesOrNoQuestion(welcomeMsg) == "y"
	if createNewAccount {
		r.createAccount()
	}
}

func (r *repl) createAccount() {
	generatePassphrase := yesOrNoQuestion(generateMsg) == "y"
	/*accountInfo := prompt.Input(prefix+accountInfoMsg,
	emptyComplete,
	prompt.OptionPrefixTextColor(prompt.LightGray))*/
	passphrase := ""
	if generatePassphrase {
		passphrase = prompt.Input(prefix+addPassphraseMsg,
			emptyComplete,
			prompt.OptionPrefixTextColor(prompt.LightGray))
	}

	err := r.client.CreateAccount(passphrase)
	if err != nil {
		log.Debug(err.Error())
		return
	}
}

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

func (r *repl) commandLineParams(idx int, input string) string {
	c := r.commands[idx]
	params := strings.Replace(input, c.text, "", -1)

	return strings.TrimSpace(params)
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
}

func (r *repl) account() {
	accountID := r.commandLineParams(3, r.input)

	if accountID != "" {
		r.client.AccountInfo(accountID)
	} else {
		accountID = inputNotBlank(getAccountInfoMsg)
		if acct := r.client.LocalAccount(); acct == nil &&
			yesOrNoQuestion(accountNotFoundoMsg) == "y" {
			r.createAccount()
		} else {
			acc, err := r.client.AccountInfo("1")
			if err != nil {
				fmt.Println(printPrefix, "Error", err.Error())
				return
			}
			fmt.Println(printPrefix, "Balance:", acc.Balance)
			fmt.Println(printPrefix, "Nonce:", acc.Nonce)
		}
	}
}

func (r *repl) transferCoins() {
	accountID := ""
	passphrase := ""

	fmt.Println(printPrefix, initialTransferMsg)

	acct := r.client.LocalAccount()
	if acct == nil {
		accountCommand := r.commands[3]

		// executing account command to create a local account
		r.executor(accountCommand.text)
		return
	}

	accountID = acct.PubKey.String()
	msg := fmt.Sprintf(transferFromLocalAccountMsg, accountID)
	isTransferFromLocal := yesOrNoQuestion(msg) == "y"

	if !isTransferFromLocal {
		accountID = inputNotBlank(transferFromAccountMsg)
	}

	destinationAccountID := inputNotBlank(transferToAccountMsg)
	amount := inputNotBlank(amountToTransferMsg)

	if !r.client.IsAccountUnLock(accountID) {
		passphrase = inputNotBlank(accountPassphrase)
	}

	acc, err := r.client.AccountInfo(accountID)
	if err != nil {
		log.Error("can't get client info")
		return
	}

	fmt.Println(printPrefix, "Transaction summary:")
	fmt.Println(printPrefix, "From:", accountID)
	fmt.Println(printPrefix, "To:", destinationAccountID)
	fmt.Println(printPrefix, "Amount:", amount)
	fmt.Println(printPrefix, "Nonce:", acc.Nonce)

	if yesOrNoQuestion(confirmTransactionMsg) == "y" {
		err := r.client.Transfer(accountID, destinationAccountID, amount, acc.Nonce, passphrase)
		if err != nil {
			log.Info(err.Error())
			return
		}
	}
}
