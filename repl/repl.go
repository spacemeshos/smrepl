package repl

import (
	"fmt"
	"os"
	"strings"

	"github.com/spacemeshos/CLIWallet/common"
	"github.com/spacemeshos/CLIWallet/log"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/ed25519"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"google.golang.org/genproto/googleapis/rpc/status"

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

	// Local account management methods
	CreateAccount(alias string) (*common.LocalAccount, error)
	CurrentAccount() (*common.LocalAccount, error)
	SetCurrentAccount(a *common.LocalAccount) error
	ListAccounts() []string
	GetAccount(name string) (*common.LocalAccount, error)
	StoreAccounts() error

	// Local config
	ServerInfo() string

	// Node service
	NodeStatus() (*apitypes.NodeStatus, error)
	NodeInfo() (*common.NodeInfo, error)
	Echo() error

	// Mesh service
	GetMeshTransactions(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Transaction, uint32, error)
	GetMeshActivations(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Activation, uint32, error)
	GetMeshInfo() (*common.NetInfo, error)

	// Transaction service
	Transfer(recipient gosmtypes.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (*apitypes.TransactionState, error)
	TransactionState(txId []byte, includeTx bool) (*apitypes.TransactionState, *apitypes.Transaction, error)

	// Smesher service
	GetSmesherId() ([]byte, error)
	IsSmeshing() (bool, error)
	StartSmeshing(address gosmtypes.Address, dataDir string, dataSizeBytes uint64) (*status.Status, error)
	StopSmeshing(deleteFiles bool) (*status.Status, error)
	GetCoinbase() (*gosmtypes.Address, error)
	SetCoinbase(coinbase gosmtypes.Address) (*status.Status, error)

	// debug service
	DebugAllAccounts() ([]*apitypes.Account, error)

	// global state service
	AccountState(address gosmtypes.Address) (*apitypes.Account, error)
	AccountRewards(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Reward, uint32, error)
	AccountTransactionsReceipts(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.TransactionReceipt, uint32, error)
	GlobalStateHash() (*apitypes.GlobalStateHash, error)
	SmesherRewards(smesherId []byte, offset uint32, maxResults uint32) ([]*apitypes.Reward, uint32, error)
}

func (r *repl) initializeCommands() {
	r.commands = []command{
		// account commands
		{"new", "Create a new account (key pair) and set as current", r.createAccount},
		{"set", "Set one of the previously created accounts as current", r.chooseAccount},
		{"info", "Display the current account info", r.printAccountInfo},
		{"rewards", "Display all rewards awarded to the current account", r.printLocalAccountRewards},
		{"sign", "Sign a hex message with the current account private key", r.sign},
		{"text-sign", "Sign a text message with the current account private key", r.textsign},

		{"any-rewards", "Display all rewards for any account", r.printAnyAccountRewards},

		// activations where this account is coinbase

		// transactions
		{"send-coin", "Transfer coins from current account to another account", r.submitCoinTransaction},
		{"tx-status", "Display a transaction status", r.printTransactionStatus},
		{"txs", "Display all outgoing and incoming transactions for the current account that are on the mesh", r.printAccountTransactions},

		// printing status and state of things
		{"node", "Display node status", r.nodeInfo},
		{"net", "Display network information", r.printMeshInfo},
		{"global-state", "Display the most recent network global state", r.printGlobalState},

		// smeshing - rewards ops
		{"print-rewards-account", "Display the currently set smesher's rewards account", r.printCoinbase},
		{"set-rewards-account", "Set current account as the node smesher's rewards account", r.setCoinbase},
		{"smesher-rewards", "Display rewards for a smesher", r.printSmesherRewards},

		// smeshing - smesher ops
		{"smesher-id", "Display the smesher's current smesher id", r.printSmesherId},
		{"start-smeshing", "Start smeshing using the current account as the rewards account", r.startSmeshing},
		{"stop-smeshing", "Stop smeshing", r.stopSmeshing},

		{"is-smeshing", "Display the proof of space status", r.printIsSmeshing},
		{"smeshing-status", "Display smeshing status", r.printSmeshingStatus},

		{"post-status", "Display the proof of space status", r.printPostStatus},
		{"post-providers", "Display the available proof of space providers", r.printPostProviders},

		// debug commands
		{"dbg-all-accounts", "Display all mesh accounts", r.printAllAccounts},

		{"quit", "Quit the CLI", r.quit},
	}
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

func (r *repl) commandLineParams(idx int, input string) string {
	c := r.commands[idx]
	params := strings.Replace(input, c.text, "", -1)

	return strings.TrimSpace(params)
}

func (r *repl) firstTime() {
	fmt.Print(printPrefix, splash)

	// TODO: change this is to use the health service when it is ready
	_, err := r.client.GetMeshInfo()
	if err != nil {
		log.Error("Failed to connect to mesh service at %v: %v", r.client.ServerInfo(), err)
		r.quit()
	}

	fmt.Println("Welcome to Spacemesh. Connected to api server at", r.client.ServerInfo())
	r.printMeshInfo()
}

func (r *repl) quit() {
	os.Exit(0)
}
