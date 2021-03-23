package repl

import (
	"encoding/hex"
	"fmt"

	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/common/util"

	"github.com/spacemeshos/CLIWallet/log"
)

// printRewards prints all rewards awarded to an account
func (r *repl) printRewards(address gosmtypes.Address) {
	// todo: request offset and total from user
	rewards, total, err := r.client.AccountRewards(address, 0, 0)
	if err != nil {
		log.Error("failed to get rewards: %v", err)
		return
	}

	fmt.Println(printPrefix, fmt.Sprintf("Total rewards: %d", total))
	for _, r := range rewards {
		printReward(r)
		fmt.Println(printPrefix, "-----")
	}
}

// printAccountRewards prints all rewards awarded to an account
func (r *repl) printAccountRewards() {
	addrStr := inputNotBlank(enterAddressMsg)
	addr := gosmtypes.HexToAddress(addrStr)
	r.printRewards(addr)
}

// printGlobalState prints the current global state
func (r *repl) printGlobalState() {
	resp, err := r.client.GlobalStateHash()
	if err != nil {
		log.Error("failed to get global state: %v", err)
		return
	}

	fmt.Println(printPrefix, "Hash:", "0x"+hex.EncodeToString(resp.RootHash))
	fmt.Println(printPrefix, "Layer:", resp.Layer.Number)
}

// printAccountState prints an account's global state
func (r *repl) printAccountState() {
	addressStr := inputNotBlank(enterAddressMsg)
	address := gosmtypes.BytesToAddress(util.FromHex(addressStr))
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

	fmt.Println(printPrefix, "Address:", address.String())
	fmt.Println(printPrefix, "Balance:", coinAmount(currBalance)) // currBalance, coinUnitName)
	fmt.Println(printPrefix, "Nonce:", state.StateCurrent.Counter)
	fmt.Println(printPrefix, "Projected Balance:", coinAmount(projectedBalance)) // projectedBalance, coinUnitName)
	fmt.Println(printPrefix, "Projected Nonce:", state.StateProjected.Counter)
	fmt.Println(printPrefix, "Projected state includes all pending transactions that haven't been added to the mesh yet.")
}
