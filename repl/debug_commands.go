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

	for _, a := range accounts {
		fmt.Println(printPrefix, "Address:", gosmtypes.BytesToAddress(a.AccountId.Address).String())
		fmt.Println(printPrefix, "Balance:", a.StateCurrent.Balance.Value, coinUnitName)
		fmt.Println(printPrefix, "Nonce:", a.StateCurrent.Counter)
		fmt.Println(printPrefix, "-----")
	}
}
