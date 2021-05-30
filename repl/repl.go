package repl

import (
	"fmt"
	"os"

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
	commands   []command
	client     Client
	clientOpen bool
	input      string
}

// Client interface to REPL clients.
type Client interface {
	WalletInfo()
	IsOpen() bool
	OpenWallet() bool
	NewWallet() bool
	CloseWallet()

	// Local account management methods
	CreateAccount(alias string) (*common.LocalAccount, error)
	CurrentAccount() (*common.LocalAccount, error)
	SetCurrentAccount(accountNumber int) error
	ListAccounts() ([]string, error)
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
	SmeshingStatus() (*apitypes.SmeshingStatusResponse, error)
	StartSmeshing(request *apitypes.StartSmeshingRequest) (*status.Status, error)
	StopSmeshing(deleteFiles bool) (*status.Status, error)
	GetPostComputeProviders(benchmark bool) ([]*apitypes.PostComputeProvider, error)
	GetSmesherId() ([]byte, error)
	GetRewardsAddress() (*gosmtypes.Address, error)
	SetRewardsAddress(coinbase gosmtypes.Address) (*status.Status, error)
	Config() (*apitypes.ConfigResponse, error)
	PostDataCreationProgressStream() (apitypes.SmesherService_PostDataCreationProgressStreamClient, error)

	// debug service
	DebugAllAccounts() ([]*apitypes.Account, error)

	// global state service
	AccountState(address gosmtypes.Address) (*apitypes.Account, error)
	AccountRewards(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Reward, uint32, error)
	AccountRewardsStream(address gosmtypes.Address) (apitypes.GlobalStateService_AccountDataStreamClient, error)
	AccountUpdatesStream(address gosmtypes.Address) (apitypes.GlobalStateService_AccountDataStreamClient, error)
	AccountTransactionsReceipts(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.TransactionReceipt, uint32, error)
	GlobalStateHash() (*apitypes.GlobalStateHash, error)
	SmesherRewards(smesherId []byte, offset uint32, maxResults uint32) ([]*apitypes.Reward, uint32, error)
}

func (r *repl) initializeCommands() {
	accountCommands := []command{
		// wallets
		{"wallet-open", "Open a wallet", r.openWallet},
		{"wallet-create", "Create a wallet", r.createWallet},
	}
	if r.clientOpen {
		accountCommands = []command{
			// local wallet account commands
			{"wallet-info", "Display wallet info", r.walletInfo},
			{"wallet-close", "Close current wallet", r.closeWallet},

			{"account-new", "Create a new account (key pair) and set as current", r.createAccount},
			{"account-set", "Set one of the previously created accounts as current", r.chooseAccount},
			{"account-info", "Display the current account info", r.printAccountInfo},
			{"account-rewards", "Display all rewards awarded to the current account", r.printLocalAccountRewards},
			{"account-sign", "Sign a hex message with the current account private key", r.sign},
			{"account-text-sign", "Sign a text message with the current account private key", r.signText},
			{"account-txs", "Display all outgoing and incoming transactions for the current account that are on the mesh", r.printCurrAccountMeshTransactions},
			{"account-send-coin", "Transfer coins from current account to another account", r.submitCoinTransaction},
		}
	}

	otherCommands := []command{

		// Misc entities status
		{"status-node", "Display node status", r.nodeInfo},
		{"status-net", "Display network information", r.printMeshInfo},
		{"status-tx", "Display a transaction status", r.printTransactionStatus},

		// global state
		{"state-account", "Display an account balance and nonce", r.printAccountState},
		{"state-account-txs", "Display account transactions in global state", r.printAccountState},

		{"state-rewards", "Display an account rewards ", r.printAccountRewards},

		// global state streams
		{"state-stream-rewards", "Stream new rewards for an account", r.printAccountRewardsStream},
		{"state-stream-account", "Stream account updates", r.printAccountUpdatesStream},

		{"state-smesher-rewards", "Display smesher rewards", r.printSmesherRewards},
		{"state-global", "Display the most recent network global state", r.printGlobalState},

		// mesh
		{"mesh-transactions", "Display mesh transaction for an account", r.printMeshTransactions},

		// smeshing - smesher ops
		{"smesher-id", "Display current smesher id", r.printSmesherId},
		{"smesher-rewards-address", "Display current smesher rewards address", r.printRewardsAddress},
		{"smesher-set-rewards-address", "Set the smesher rewards address", r.setRewardsAddress},
		{"smesher-rewards", "Display current smesher rewards", r.printCurrentSmesherRewards},
		{"smesher-status", "Display smesher status", r.printSmeshingStatus},
		{"smesher-stop", "Stop smeshing", r.stopSmeshing},

		// proof of space methods
		{"pos-providers", "Display the available proof of space providers", r.printPosProviders},
		{"pos-setup", "Set up (or change) smesher proof of space data", r.setupPos},
		{"pos-stream-progress", "Stream data creation progress", r.printPostDataCreationProgress},

		// debug commands
		{"dbg-all-accounts", "Display all global state accounts", r.printAllAccounts},

		{"quit", "Quit this app", r.quit},
	}
	r.commands = append(accountCommands, otherCommands...)
}

// Start starts the REPL
func Start(c Client) {

	// init logging system
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(printPrefix, "Aborting. Can't get current dir. Your system is high.", err)
		return
	}

	log.InitSpacemeshLoggingSystem(path, "log.txt")

	log.Info("new session started")

	if !TestMode {
		r := &repl{client: c}
		r.clientOpen = c.IsOpen()
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

func (r *repl) firstTime() {
	fmt.Print(splash)

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
