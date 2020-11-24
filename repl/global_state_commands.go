package repl

import (
	"encoding/hex"
	"fmt"

	"github.com/spacemeshos/CLIWallet/log"
)

// Outputs the current global state
func (r *repl) printGlobalState() {

	resp, err := r.client.GlobalStateHash()
	if err != nil {
		log.Error("failed to get global state: %v", err)
		return
	}

	fmt.Println(printPrefix, "Hash:", "0x"+hex.EncodeToString(resp.RootHash))
	fmt.Println(printPrefix, "Layer:", resp.Layer.Number)

}
