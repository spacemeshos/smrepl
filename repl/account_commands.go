package repl

import (
	"encoding/hex"
	"fmt"

	"github.com/spacemeshos/CLIWallet/common"
	"github.com/spacemeshos/CLIWallet/log"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/ed25519"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

func (r *repl) chooseAccount() {
	accs := r.client.ListAccounts()
	if len(accs) == 0 {
		r.createAccount()
		return
	}

	fmt.Println(printPrefix, "Choose an account to load:")
	accName := multipleChoice(accs)
	account, err := r.client.GetAccount(accName)
	if err != nil {
		panic("wtf")
	}
	fmt.Printf("%s Loaded account alias: `%s`, address: %s \n", printPrefix, account.Name, account.Address().String())

	r.client.SetCurrentAccount(account)
}

func (r *repl) createAccount() {
	fmt.Println(printPrefix, "Create a new account")
	alias := inputNotBlank(createAccountMsg)

	ac, err := r.client.CreateAccount(alias)
	if err != nil {
		log.Error("Failed to create a new account: %v", err)
		return
	}
	err = r.client.StoreAccounts()
	if err != nil {
		log.Error("Failed to save the new account: %v", err)
		return
	}

	fmt.Printf("%s Created account: %s, address: %s \n", printPrefix, ac.Name, ac.Address().String())

	r.client.SetCurrentAccount(ac)
}

// print account info from global state
func (r *repl) printAccountInfo() {
	acc, err := r.getCurrent()
	if err != nil {
		log.Error("failed to get account", err)
		return
	}

	address := gosmtypes.BytesToAddress(acc.PubKey)

	state, err := r.client.AccountState(address)
	if err != nil {
		log.Error("failed to get account info: %v", err)
		return
	}

	currBalance := uint64(0)
	if state.StateCurrent.Balance != nil {
		currBalance = state.StateCurrent.Balance.Value
	}

	projectedBalance := uint64(0)
	if state.StateProjected.Balance != nil {
		projectedBalance = state.StateProjected.Balance.Value
	}

	fmt.Println(printPrefix, "Local alias:", acc.Name)
	fmt.Println(printPrefix, "Address:", address.String())
	fmt.Println(printPrefix, "Balance:", currBalance, coinUnitName)
	fmt.Println(printPrefix, "Nonce:", state.StateCurrent.Counter)
	fmt.Println(printPrefix, "Projected Balance:", projectedBalance, coinUnitName)
	fmt.Println(printPrefix, "Projected Nonce:", state.StateProjected.Counter)
	fmt.Println(printPrefix, "Projected account state includes all pending transactions that haven't been added to the mesh yet.")
	fmt.Println(printPrefix, fmt.Sprintf("Public key: 0x%s", hex.EncodeToString(acc.PubKey)))
	fmt.Println(printPrefix, fmt.Sprintf("Private key: 0x%s", hex.EncodeToString(acc.PrivKey)))
}

// printAccountRewards prints all rewards awarded to an account
func (r *repl) printRewards(address gosmtypes.Address) {
	// todo: request offset and total from user
	rewards, total, err := r.client.AccountRewards(address, 0, 10000)
	if err != nil {
		log.Error("failed to list transactions: %v", err)
		return
	}

	fmt.Println(printPrefix, fmt.Sprintf("Total rewards: %d", total))
	for _, r := range rewards {
		printReward(r)
		fmt.Println(printPrefix, "-----")
	}
}

// printAccountRewards prints all rewards awarded to an account
func (r *repl) printLocalAccountRewards() {
	acc, err := r.getCurrent()
	if err != nil {
		log.Error("failed to get account", err)
		return
	}
	r.printRewards(acc.Address())
}

// printAccountRewards prints all rewards awarded to an account
func (r *repl) printAnyAccountRewards() {
	addrStr := inputNotBlank(enterAddressMsg)
	addr := gosmtypes.HexToAddress(addrStr)

	r.printRewards(addr)
}

func printReward(r *apitypes.Reward) {
	fmt.Println(printPrefix, "Rewarded on layer:", r.Layer.Number)
	//fmt.Println(printPrefix, "Rewarded for layer:", r.LayerComputed.Number)
	fmt.Println(printPrefix, "Layer reward", r.LayerReward.Value, coinUnitName)
	fmt.Println(printPrefix, "Transaction fees", r.Total.Value-r.LayerReward.Value, coinUnitName)
	fmt.Println(printPrefix, "Total reward", r.Total.Value, coinUnitName)
	//fmt.Println(printPrefix, "Smesher id", "0x"+hex.EncodeToString(r.Smesher.Id))
	fmt.Println(printPrefix, "Rewards account:", gosmtypes.BytesToAddress(r.Coinbase.Address).String())
}

func (r *repl) getCurrent() (acc *common.LocalAccount, err error) {
	acc, err = r.client.CurrentAccount()
	if err != nil {
		r.chooseAccount()
		acc, err = r.client.CurrentAccount()
	}
	return
}

func (r *repl) sign() {
	acc, err := r.getCurrent()
	if err != nil {
		log.Error("failed to get account", err)
		return
	}

	msgStr := inputNotBlank(msgSignMsg)
	msg, err := hex.DecodeString(msgStr)
	if err != nil {
		log.Error("failed to decode msg hex string: %v", err)
		return
	}

	signature := ed25519.Sign2(acc.PrivKey, msg)

	fmt.Println(printPrefix, fmt.Sprintf("signature (in hex): %x", signature))
}

func (r *repl) textsign() {
	acc, err := r.getCurrent()
	if err != nil {
		log.Error("failed to get account", err)
		return
	}

	msg := inputNotBlank(msgTextSignMsg)
	signature := ed25519.Sign2(acc.PrivKey, []byte(msg))

	fmt.Println(printPrefix, fmt.Sprintf("signature (in hex): %x", signature))
}
