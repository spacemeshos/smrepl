package repl

import (
	"strconv"

	"github.com/spacemeshos/CLIWallet/log"
)

func (r *repl) smesh() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	datadir := inputNotBlank(smeshingDatadirMsg)

	spaceStr := inputNotBlank(smeshingSpaceAllocationMsg)
	space, err := strconv.ParseUint(spaceStr, 10, 32)
	if err != nil {
		log.Error("failed to parse: %v", err)
		return
	}

	if err := r.client.Smesh(datadir, uint(space)<<30, acc.Address().String()); err != nil {
		log.Error("failed to start smeshing: %v", err)
		return
	}
}
