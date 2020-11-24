package repl

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spacemeshos/CLIWallet/log"
)

/*
{"get-rewards-account", "Get current account as the node smesher's rewards account", r.getCoinbase},
		{"set-rewards-account", "Set current account as the node smesher's rewards account", r.setCoinbase},
		{"get-smesher-id", "Get the node smesher's current rewards account", r.getSmesherId},
		{"set-rewards-account", "Set current account as the node smesher's rewards account", r.setCoinbase},
		{"is-smeshing", "Set current account as the node smesher's rewards account", r.isSmeshing},
*/

func (r *repl) startSmeshing() {
	addr := r.client.CurrentAccount()
	if addr == nil {
		r.chooseAccount()
		addr = r.client.CurrentAccount()
	}

	datadir := inputNotBlank(smeshingDatadirMsg)

	spaceStr := inputNotBlank(smeshingSpaceAllocationMsg)
	dataSize, err := strconv.ParseUint(spaceStr, 10, 64)
	if err != nil {
		log.Error("failed to parse: %v", err)
		return
	}

	resp, err := r.client.StartSmeshing(addr.Address(), datadir, dataSize)

	if err != nil {
		log.Error("failed to start smeshing: %v", err)
		return
	}

	if resp.Code != 0 {
		log.Error("failed to start smeshing. Response status: %d", resp.Code)
		return
	}

	fmt.Println(printPrefix, "Smeshing started")

}

func (r *repl) stopSmeshing() {

	deleteData := yesOrNoQuestion(confirmDeleteDataMsg) == "y"

	resp, err := r.client.StopSmeshing(deleteData)

	if err != nil {
		log.Error("failed to stop smeshing: %v", err)
		return
	}

	if resp.Code != 0 {
		log.Error("failed to stop smeshing. Response status: %d", resp.Code)
		return
	}

	fmt.Println(printPrefix, "Smeshing started")

}

func (r *repl) printSmeshingStatus() {
	isSmeshing, err := r.client.IsSmeshing()

	if err != nil {
		log.Error("failed to get smeshing status: %v", err)
		return
	}

	if isSmeshing {
		fmt.Println(printPrefix, "Smeshing is currently on")
	} else {
		fmt.Println(printPrefix, "Smeshing is off")
	}
}

func (r *repl) printCoinbase() {
	if resp, err := r.client.GetCoinbase(); err != nil {
		log.Error("failed to get rewards address: %v", err)
	} else {
		fmt.Println(printPrefix, "Rewards address is:", resp.String())
	}
}

func (r *repl) printSmesherId() {
	if resp, err := r.client.GetSmesherId(); err != nil {
		log.Error("failed to get smesher id: %v", err)
	} else {
		fmt.Println(printPrefix, "Smesher id:", "0x"+hex.EncodeToString(resp))
	}
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
