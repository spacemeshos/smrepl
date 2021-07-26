package repl

import (
	"encoding/hex"
	"fmt"
	"github.com/spacemeshos/ed25519"
	"strconv"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/util"

	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/smrepl/log"
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

// Submit a transaction using the current set user account
func (r *repl) submitCoinTransactionWithCurrentAccount() {

	fmt.Println(printPrefix, initialTransferMsg)

	if !r.canSubmitTransactions() {
		fmt.Println(printPrefix, "Can't submit a new transaction right now because the node is not synced. Please try again later")
		return
	}

	account, err := r.getCurrent()
	if err != nil {
		log.Error("failed to get current account", err)
		return
	}
	srcAddress := gosmtypes.BytesToAddress(account.PubKey)
	acctState, err := r.client.AccountState(srcAddress)
	if err != nil {
		log.Error("failed to get account info: %v", err)
		return
	}
	counter := acctState.StateProjected.Counter

	r.submitCoinTransaction(srcAddress, counter, account.PrivKey)
}

// Submit a coin transaction from any account
func (r *repl) submitCoinTransactionAnyAccount() {

	fmt.Println(printPrefix, transferMsgAnyAccount)

	if !r.canSubmitTransactions() {
		fmt.Println(printPrefix, "Can't submit a new transaction right now because the node is not synced. Please try again later")
		return
	}

	addrStr := inputNotBlank("Enter account address")
	srcAddress, err := gosmtypes.StringToAddress(addrStr)
	if err != nil {
		log.Error("invalid input address", err)
		return
	}
	privateKeyStr := inputNotBlank("Enter account private key")
	privateKeyBytes := util.FromHex(privateKeyStr)
	privateKey := ed25519.PrivateKey(privateKeyBytes)

	counterStr := inputNotBlank("Enter account counter")
	counter, err := strconv.ParseUint(counterStr, 10, 64)
	if err != nil {
		log.Error("invalid counter. Must be 0 or larger integer", err)
		return
	}

	r.submitCoinTransaction(srcAddress, counter, privateKey)
}

// Submit a new transaction using the provided sender data
func (r *repl) submitCoinTransaction(srcAddress gosmtypes.Address, counter uint64, srcPrivateKey ed25519.PrivateKey) {

	destAddressStr := inputNotBlank(destAddressMsg)
	destAddress, err := gosmtypes.StringToAddress(destAddressStr)
	if err != nil {
		log.Error("invalid address")
		return
	}
	amountStr := inputNotBlank(amountToTransferMsg)
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		log.Error("invalid amount. Must be a non-negative integer number")
	}

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
	fmt.Println(printPrefix, "Nonce: ", counter)

	if yesOrNoQuestion(confirmTransactionMsg) == "y" {
		txState, err := r.client.Transfer(destAddress, counter, amount, gas, 100, srcPrivateKey)
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
