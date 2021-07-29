package repl

import (
	"encoding/base64"
	"fmt"

	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/common/util"
	"github.com/spacemeshos/smrepl/log"
)

func (r *repl) printAllAccounts() {
	accounts, err := r.client.DebugAllAccounts()
	if err != nil {
		log.Error("failed to get debug all accounts: %v", err)
		return
	}

	for i, a := range accounts {

		fmt.Println("Address:", gosmtypes.BytesToAddress(a.AccountId.Address).String())

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

// hexTobase64 returns the standard (= padding) base64 encoded string for a hex string
func (r *repl) hexToBase64() {
	data := inputNotBlank("Enter a hex string: ")
	fmt.Println(base64.StdEncoding.EncodeToString(util.FromHex(data)))
}
