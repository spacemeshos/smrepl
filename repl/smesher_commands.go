package repl

import (
	"encoding/hex"
	"fmt"
	"strconv"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"

	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"

	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/go-spacemesh/common/util"
)

// number of bytes in 1 GiB
const GIB uint64 = 1_262_485_504

func (r *repl) printSmeshingStatus() {
	isSmeshing, err := r.client.IsSmeshing()
	if err != nil {
		log.Error("failed to query for smeshing status: %v", err)
		return
	}

	if isSmeshing {
		fmt.Println(printPrefix, "node is smeshing")
	} else {
		fmt.Println(printPrefix, "node is not smeshing")
	}
}

/// setupPos start an interactive proof of space data creation process
func (r *repl) setupPos() {

	addrStr := inputNotBlank(enterRewardsAddress)
	addr := gosmtypes.HexToAddress(addrStr)
	dataDir := inputNotBlank(posDataDirMsg)
	spaceGiBStr := inputNotBlank(posSizeMsg)
	dataSizeGiB, err := strconv.ParseUint(spaceGiBStr, 10, 64)
	if err != nil {
		log.Error("failed to parse your input: %v", err)
		return
	}

	numLabels := dataSizeGiB / GIB

	providerIdStr := inputNotBlank(posProviderMsg)
	providerId, err := strconv.ParseUint(providerIdStr, 10, 32)
	if err != nil {
		log.Error("failed to parse your input: %v", err)
		return
	}

	req := &apitypes.StartSmeshingRequest{}
	req.Coinbase = &apitypes.AccountId{Address: addr.Bytes()}
	req.Opts = &apitypes.PostInitOpts{
		DataDir:           dataDir,
		NumLabels:         numLabels,
		NumFiles:          1,
		ComputeProviderId: uint32(providerId),
		Throttle:          false,
	}

	resp, err := r.client.StartSmeshing(req)

	if err != nil {
		log.Error("failed to start smeshing: %v", err)
		return
	}

	if resp.Code != 0 {
		log.Error("failed to start smeshing. Response status: %d", resp.Code)
		return
	}

	fmt.Println(printPrefix, "Smeshing started. Add the following to your node's config file so it will continue smeshing after you restart it")
	fmt.Println(printPrefix, "todo: Json to add to node cofig file here")

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

	fmt.Println(printPrefix, "Smeshing stopped. Don't forget to remove smeshing related data from your node's config file or startup flags so it won't start smeshing after you restart it")

}

var ComputeApiClass_name = map[int32]string{
	0: "COMPUTE_API_CLASS_UNSPECIFIED",
	1: "COMPUTE_API_CLASS_CPU",
	2: "COMPUTE_API_CLASS_CUDA",
	3: "COMPUTE_API_CLASS_VULKAN",
}

/// setupProofOfSpace prints the available proof of space compute providers
func (r *repl) printPosProviders() {

	providers, err := r.client.GetPostComputeProviders()
	if err != nil {
		log.Error("failed to get compute providers: %v", err)
		return
	}

	if len(providers) == 0 {
		fmt.Println(printPrefix, "No supported compute providers found")
		return
	}

	fmt.Println(printPrefix, "Supported providers on your system:")

	for i, p := range providers {
		if i != 0 {
			fmt.Println(printPrefix, "---_-")
		}
		fmt.Println(printPrefix, "Provider id:", p.GetId())
		fmt.Println(printPrefix, "Model:", p.GetModel())
		fmt.Println(printPrefix, "Compute api:", ComputeApiClass_name[int32(p.GetComputeApi())])
		fmt.Println(printPrefix, "Performance:", p.GetPerformance())
	}
}

func (r *repl) printSmesherId() {
	if resp, err := r.client.GetSmesherId(); err != nil {
		log.Error("failed to get smesher id: %v", err)
	} else {
		fmt.Println(printPrefix, "Smesher id:", "0x"+hex.EncodeToString(resp))
	}
}

func (r *repl) printRewardsAddress() {
	if resp, err := r.client.GetRewardsAddress(); err != nil {
		log.Error("failed to get rewards address: %v", err)
	} else {
		fmt.Println(printPrefix, "Rewards address is:", resp.String())
	}
}

// setRewardsAddress sets the smesher's reward address to a user provider address
func (r *repl) setRewardsAddress() {
	addrStr := inputNotBlank(enterAddressMsg)
	addr := gosmtypes.HexToAddress(addrStr)

	resp, err := r.client.SetRewardsAddress(addr)

	if err != nil {
		log.Error("failed to set rewards address: %v", err)
		return
	}

	if resp.Code == 0 {
		fmt.Println(printPrefix, "Rewards address set to:", addr.String())
	} else {
		// todo: what are the possible non-zero status codes here?
		fmt.Println(printPrefix, fmt.Sprintf("Response status code: %d", resp.Code))
	}
}

////////// The following methods use the global state service and not the smesher service

// printSmesherRewards prints all rewards awarded to a smesher identified by an id
func (r *repl) printSmesherRewards() {

	smesherIdStr := inputNotBlank(smesherIdMsg)
	smesherId := util.FromHex(smesherIdStr)

	// todo: request offset and total from user
	rewards, total, err := r.client.SmesherRewards(smesherId, 0, 0)
	if err != nil {
		log.Error("failed to get rewards: %v", err)
		return
	}

	fmt.Println(printPrefix, fmt.Sprintf("Total rewards: %d", total))
	for i, r := range rewards {
		if i != 0 {
			fmt.Println(printPrefix, "-----")
		}
		printReward(r)
	}
}

// printSmesherRewards prints all rewards awarded to a smesher identified by an id
func (r *repl) printCurrentSmesherRewards() {
	if smesherId, err := r.client.GetSmesherId(); err != nil {
		log.Error("failed to get smesher id: %v", err)
	} else {

		fmt.Println(printPrefix, "Smesher id:", "0x"+hex.EncodeToString(smesherId))

		// todo: request offset and total from user
		rewards, total, err := r.client.SmesherRewards(smesherId, 0, 10000)
		if err != nil {
			log.Error("failed to get rewards: %v", err)
			return
		}

		fmt.Println(printPrefix, fmt.Sprintf("Total rewards: %d", total))
		for i, r := range rewards {
			if i != 0 {
				fmt.Println(printPrefix, "-----")
			}
			printReward(r)
		}
	}
}
