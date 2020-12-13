package repl

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/go-spacemesh/common/util"
)

// printSmesherRewards prints all rewards awarded to a smesher identified by an id
func (r *repl) printSmesherRewards() {

	smesherIdStr := inputNotBlank(smesherIdMsg)
	smesherId := util.FromHex(smesherIdStr)

	// todo: request offset and total from user
	rewards, total, err := r.client.SmesherRewards(smesherId, 0, 10000)
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

func (r *repl) startSmeshing() {
	addr, err := r.getCurrent()
	if err != nil {
		log.Error("failed to get account", err)
		return
	}

	dataDir := inputNotBlank(smeshingDatadirMsg)

	spaceGBStr := inputNotBlank(smeshingSpaceAllocationMsg)
	dataSizeGB, err := strconv.ParseUint(spaceGBStr, 10, 64)
	if err != nil {
		log.Error("failed to parse: %v", err)
		return
	}

	resp, err := r.client.StartSmeshing(addr.Address(), dataDir, uint64(dataSizeGB<<20))

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

func (r *repl) printPostStatus() {

}

func (r *repl) printPostProviders() {

}

func (r *repl) printSmeshingStatus() {

}

func (r *repl) printIsSmeshing() {
	isSmeshing, err := r.client.IsSmeshing()

	if err != nil {
		log.Error("failed to get smeshing status: %v", err)
		return
	}

	if isSmeshing {
		fmt.Println(printPrefix, "Smeshing is on")
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
	acc, err := r.getCurrent()
	if err != nil {
		log.Error("failed to get account", err)
		return
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
