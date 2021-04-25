package repl

import (
	"fmt"

	"github.com/spacemeshos/CLIWallet/log"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

func (r *repl) printAllAccounts() {
	accounts, err := r.client.DebugAllAccounts()
	if err != nil {
		log.Error("failed to get debug all accounts: %v", err)
		return
	}

	for i, a := range accounts {

		fmt.Println(printPrefix, "Address:", gosmtypes.BytesToAddress(a.AccountId.Address).String())

		balance := uint64(0)
		if a.StateCurrent.Balance != nil {
			balance = a.StateCurrent.Balance.Value
		}

		if i != 0 {
			fmt.Println("-----")
		}

		fmt.Println("Balance:", balance, coinUnitName)
		fmt.Println("Nonce:", a.StateCurrent.Counter)
	}
}
