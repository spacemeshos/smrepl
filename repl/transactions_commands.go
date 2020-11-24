package repl

import (
	"encoding/hex"
	"fmt"
	"strconv"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/util"

	"github.com/spacemeshos/CLIWallet/log"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

var transactionStateDisStringsMap = map[int32]string{
	0: "Unspecified state",
	1: "Rejected",
	2: "Insufficient funds",
	3: "Conflicting",
	4: "Submitted to the network",
	5: "On the mesh but not yet processed",
	6: "Processed",
}

// Print a transaction status
func (r *repl) printTransactionStatus() {
	txIdStr := inputNotBlank(txIdMsg)
	txId := util.FromHex(txIdStr)
	txState, tx, err := r.client.TransactionState(txId, true)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if txState != nil {
		txStateDispString := transactionStateDisStringsMap[int32(txState.State.Number())]
		fmt.Println(printPrefix, "State:", txStateDispString)
	} else {
		fmt.Println(printPrefix, "Unknown transaction state")
	}

	if tx != nil {
		printTransaction(tx)
	} else {
		fmt.Println(printPrefix, "Unknown transaction")
	}
}

// canSubmitTransactions returns true if the node is accepting transactions.
// todo: this should move to a method in the transactions service.
func (r *repl) canSubmitTransactions() bool {

	status, err := r.client.NodeStatus()
	if err != nil {
		log.Error("failed to get node status: %v", err)
		return false
	}

	// for now, we allow to submit txs if the node is synced
	return status.IsSynced //&& status.TopLayer.Number > minVerifiedLayer
}

func (r *repl) submitCoinTransaction() {

	if !r.canSubmitTransactions() {
		fmt.Println(printPrefix, "Can't submit a new transaction. Please try again later")
		return
	}
	fmt.Println(printPrefix, initialTransferMsg)
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	srcAddress := gosmtypes.BytesToAddress(acc.PubKey)
	acctState, err := r.client.AccountState(srcAddress)
	if err != nil {
		log.Error("failed to get account info: %v", err)
		return
	}

	destAddressStr := inputNotBlank(destAddressMsg)
	destAddress := gosmtypes.HexToAddress(destAddressStr)

	amountStr := inputNotBlank(amountToTransferMsg)

	gas := uint64(1)
	if yesOrNoQuestion(useDefaultGasMsg) == "n" {
		gasStr := inputNotBlank(enterGasPrice)
		gas, err = strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			log.Error("invalid transaction fee", err)
			return
		}
	}

	fmt.Println(printPrefix, "New transaction summary:")
	fmt.Println(printPrefix, "From:  ", srcAddress.String())
	fmt.Println(printPrefix, "To:    ", destAddress.String())
	fmt.Println(printPrefix, "Amount:", amountStr, coinUnitName)
	fmt.Println(printPrefix, "Fee:   ", gas, coinUnitName)
	fmt.Println(printPrefix, "Nonce: ", acctState.StateProjected.Counter)

	amount, _ := strconv.ParseUint(amountStr, 10, 64)
	// todo: handle error here!

	if yesOrNoQuestion(confirmTransactionMsg) == "y" {
		txState, err := r.client.Transfer(destAddress, acctState.StateProjected.Counter, amount, gas, 100, acc.PrivKey)
		if err != nil {
			log.Error(err.Error())
			return
		}

		txStateDispString := transactionStateDisStringsMap[int32(txState.State.Number())]

		fmt.Println(printPrefix, "Transaction submitted.")
		fmt.Println(printPrefix, fmt.Sprintf("Transaction id: 0x%v", hex.EncodeToString(txState.Id.Id)))
		fmt.Println(printPrefix, "Transaction state:", txStateDispString)
	}
}

func (r *repl) printAccountTransactions() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	// todo: request offset and total from user
	txs, total, err := r.client.GetMeshTransactions(acc.Address(), 0, 1000)
	if err != nil {
		log.Error("failed to print transactions: %v", err)
		return
	}

	fmt.Println(printPrefix, fmt.Sprintf("Total mesh transactions: %d", total))
	for _, tx := range txs {
		printTransaction(tx)
		fmt.Println(printPrefix, "-----")
	}
}

// helper method - prints tx info
func printTransaction(t *apitypes.Transaction) {

	txIdStr := "0x" + util.Bytes2Hex(t.Id.Id)
	fmt.Println(printPrefix, fmt.Sprintf("Transaction id: %v", txIdStr))
	fmt.Println(printPrefix, "From:", gosmtypes.BytesToAddress(t.Sender.Address).String())

	ct := t.GetCoinTransfer()
	if ct != nil {
		fmt.Println(printPrefix, "To (coin account):", gosmtypes.BytesToAddress(ct.Receiver.Address).String())
		fmt.Println(printPrefix, "Nonce:", t.Counter)
		fmt.Println(printPrefix, "Amount:", t.Amount.Value, coinUnitName)
		fmt.Println(printPrefix, "Fee:", t.GasOffered.GasProvided, coinUnitName)
		return
	}

	sct := t.GetSmartContract()
	if sct == nil {
		log.Error("expected a smart contract transaction type")
		return
	}

	// todo: printout smart contract transaction data here
}
