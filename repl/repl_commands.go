package repl

import (
	"encoding/hex"
	"fmt"

	"github.com/spacemeshos/CLIWallet/log"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

func (r *repl) nodeInfo() {

	info, err := r.client.NodeInfo()
	if err != nil {
		log.Error("failed to get node info: %v", err)
		return
	}

	fmt.Println(printPrefix, "Version:", info.Version)
	fmt.Println(printPrefix, "Build:", info.Build)
	fmt.Println(printPrefix, "API server:", r.client.ServerUrl())

	status, err := r.client.NodeStatus()
	if err != nil {
		log.Error("failed to get node status: %v", err)
		return
	}

	fmt.Println(printPrefix, "Synced:", status.IsSynced)
	fmt.Println(printPrefix, "Synced layer:", status.SyncedLayer.Number)
	fmt.Println(printPrefix, "Current layer:", status.TopLayer.Number)
	fmt.Println(printPrefix, "Verified layer:", status.VerifiedLayer.Number)
	fmt.Println(printPrefix, "Peers:", status.ConnectedPeers)

	/*
		fmt.Println(printPrefix, "Smeshing data directory:", info.SmeshingDatadir)
		fmt.Println(printPrefix, "Smeshing status:", info.SmeshingStatus)
		fmt.Println(printPrefix, "Smeshing coinbase:", info.SmeshingCoinbase)
		fmt.Println(printPrefix, "Smeshing remaining bytes:", info.SmeshingRemainingBytes)
	*/
}

// Outputs the current global state
func (r *repl) printGlobalState() {

	resp, err := r.client.GlobalStateHash()
	if err != nil {
		log.Error("failed to get global state: %v", err)
		return
	}

	fmt.Println(printPrefix, "Hash: 0x", hex.EncodeToString(resp.RootHash))
	fmt.Println(printPrefix, "Layer:", resp.Layer.Number)

}

func (r *repl) printAllAccounts() {

	accounts, err := r.client.DebugAllAccounts()
	if err != nil {
		log.Error("failed to get debug all accounts: %v", err)
		return
	}

	for _, a := range accounts {
		fmt.Println(printPrefix, "Address:", gosmtypes.BytesToAddress(a.AccountId.Address).String())
		fmt.Println(printPrefix, "Balance:", a.StateCurrent.Balance.Value, coinUnitName)
		fmt.Println(printPrefix, "Nonce:", a.StateCurrent.Counter)
		fmt.Println(printPrefix, "-----")
	}
}

func (r *repl) printTransactionStatus() {
	txIdStr := inputNotBlank(txIdMsg)
	txId, err := hex.DecodeString(txIdStr)
	if err != nil {
		log.Error("failed to parse transaction id: %v", err)
		return
	}

	txState, tx, err := r.client.TransactionState(txId, true)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if txState != nil {
		fmt.Println(printPrefix, "State:", txState.State.Descriptor())
	} else {
		fmt.Println(printPrefix, "Unknown transaction state")
	}

	if tx != nil {
		printTransaction(tx)
	} else {
		fmt.Println(printPrefix, "Unknown transaction")
	}
}
