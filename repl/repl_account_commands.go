package repl

import (
	"encoding/hex"
	"fmt"
	"strconv"

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

	ac := r.client.CreateAccount(alias)
	err := r.client.StoreAccounts()
	if err != nil {
		log.Error("failed to create account: %v", err)
		return
	}

	fmt.Printf("%s Created account alias: `%s`, address: %s \n", printPrefix, ac.Name, ac.Address().String())
	r.client.SetCurrentAccount(ac)
}

func (r *repl) accountInfo() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	address := gosmtypes.BytesToAddress(acc.PubKey)

	state, err := r.client.AccountState(address)
	if err != nil {
		log.Error("failed to get account info: %v", err)
		state = &common.AccountState{}
	}

	fmt.Println(printPrefix, "Local alias:", acc.Name)
	fmt.Println(printPrefix, "Address:", address.String())
	fmt.Println(printPrefix, "Balance:", state.Balance, coinUnitName)
	fmt.Println(printPrefix, "Nonce:", state.Nonce)
	fmt.Println(printPrefix, fmt.Sprintf("Public key: 0x%s", hex.EncodeToString(acc.PubKey)))
	fmt.Println(printPrefix, fmt.Sprintf("Private key: 0x%s", hex.EncodeToString(acc.PrivKey)))
}

// canSubmitTransactions returns true if the node is accepting transactions.
// todo: this should move to a method in the transactions service.
func (r *repl) canSubmitTransactions() bool {

	status, err := r.client.NodeStatus()
	if err != nil {
		log.Error("failed to get node status: %v", err)
		return false
	}

	return status.IsSynced && status.TopLayer.Number > minVerifiedLayer

}

func (r *repl) submitCoinTransaction() {

	if !r.canSubmitTransactions() {
		fmt.Println(printPrefix, "Can't submit a new transaction. Please try again when node is synced and current layer is", minVerifiedLayer)
		return
	}
	fmt.Println(printPrefix, initialTransferMsg)
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	srcAddress := gosmtypes.BytesToAddress(acc.PubKey)
	info, err := r.client.AccountState(srcAddress)
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

	fmt.Println(printPrefix, "Transaction summary:")
	fmt.Println(printPrefix, "From:  ", srcAddress.String())
	fmt.Println(printPrefix, "To:    ", destAddress.String())
	fmt.Println(printPrefix, "Amount:", amountStr, coinUnitName)
	fmt.Println(printPrefix, "Fee:   ", gas, coinUnitName)
	fmt.Println(printPrefix, "Nonce: ", info.Nonce)

	amount, _ := strconv.ParseUint(amountStr, 10, 64)
	// todo: handle error here!

	if yesOrNoQuestion(confirmTransactionMsg) == "y" {
		txState, err := r.client.Transfer(destAddress, info.Nonce, amount, gas, 100, acc.PrivKey)
		if err != nil {
			log.Error(err.Error())
			return
		}

		fmt.Println(printPrefix, "Transaction submitted.")
		fmt.Println(printPrefix, fmt.Sprintf("Transaction id: 0x%v", hex.EncodeToString(txState.Id.Id)))
		fmt.Println(printPrefix, fmt.Sprintf("Transaction state: 0x%v", txState.State.String()))
	}
}

// printAccountRewards prints all rewards awarded to an account
func (r *repl) printAccountRewards() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	// todo: request offset and total from user
	rewards, total, err := r.client.AccountRewards(acc.Address(), 0, 10000)
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

func (r *repl) printAccountTransactions() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	// todo: request offset and total from user
	txs, total, err := r.client.GetMeshTransactions(acc.Address(), 0, 1000)
	if err != nil {
		log.Error("failed to list transactions: %v", err)
		return
	}

	fmt.Println(printPrefix, fmt.Sprintf("Total mesh transactions: %d", total))
	for _, tx := range txs {
		printTransaction(tx)
		fmt.Println(printPrefix, "-----")
	}
}

func printReward(r *apitypes.Reward) {

}

// helper method - prints tx info
func printTransaction(t *apitypes.Transaction) {

	fmt.Println(printPrefix, "Transaction summary:")
	fmt.Println(printPrefix, "From:", gosmtypes.BytesToAddress(t.Sender.Address).String())
	fmt.Println(printPrefix, "Amount:", t.Amount.Value, coinUnitName)
	fmt.Println(printPrefix, "Nonce:", t.Counter)

	ct := t.GetCoinTransfer()
	if ct != nil {
		fmt.Println(printPrefix, "To (coin account):", gosmtypes.BytesToAddress(ct.Receiver.Address).String())
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

func (r *repl) setCoinbase() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	resp, err := r.client.SetCoinbase(acc.Address())

	if err != nil {
		log.Error("failed to set rewards address: %v", err)
		return
	}

	if resp.Code == 0 {
		fmt.Println(printPrefix, "Rewards address set to:", acc.Address().String())
	} else {
		// todo: what are possible non-zero status codes here?
		fmt.Println(printPrefix, fmt.Sprintf("Response status code: %d", resp.Code))
	}
}

func (r *repl) sign() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
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
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	msg := inputNotBlank(msgTextSignMsg)
	signature := ed25519.Sign2(acc.PrivKey, []byte(msg))

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
