package repl

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/spacemeshos/smrepl/common"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"

	"github.com/spacemeshos/go-spacemesh/common/util"
	"github.com/spacemeshos/smrepl/log"
)

// gib is the number of bytes in 1 gibibyte (2^30 bytes)
const gib uint64 = 1073741824
const posDataFileName = "pos-data.json"

func (r *repl) printSmeshingStatus() {
	res, err := r.client.IsSmeshing()
	if err != nil {
		log.Error("failed to get smeshing status: %v", err)
		return
	}

	fmt.Printf("IsSmeshing: %v\n", res.IsSmeshing)
}

func (r *repl) printPostStatus() {
	res, err := r.client.PostStatus()
	if err != nil {
		log.Error("failed to get proof of space status: %v", err)
		return
	}

	switch res.Status.State {
	case apitypes.PoSTSetupStatus_STATE_NOT_STARTED:
		fmt.Println("Proof of space data is not setup. Enter `pos-setup` to set it up.")
		return
	case apitypes.PoSTSetupStatus_STATE_IN_PROGRESS:
		fmt.Println("⏱  Proof of space data creation is in progress.")
	case apitypes.PoSTSetupStatus_STATE_COMPLETE:
		fmt.Println("👍  Proof of space data was created and is used for smeshing.")
	case apitypes.PoSTSetupStatus_STATE_ERROR:
		fmt.Printf("⚠️  Proof of space data creation error: %v", res.Status.ErrorMessage)
	default:
		fmt.Println("printPrefix", "Unexpected api result.")
		return
	}

	cfg, err := r.client.Config()
	if err != nil {
		log.Error("failed get proof of space config from node: %v", err)
		return
	}

	unitSizeBytes := uint64(cfg.BitsPerLabel) * cfg.LabelsPerUnit / 8
	unitSizeInGiB := float32(unitSizeBytes) / float32(gib)
	opts := res.Status.Opts

	println()
	fmt.Println("Proof of space info:")
	fmt.Println("  Data dir (relative to node or absolute):", opts.DataDir)
	fmt.Println("  Date files:", opts.NumFiles)
	fmt.Println("  Compute provider id:", opts.ComputeProviderId)
	fmt.Println("  Throttle when computer is busy:", opts.Throttle)
	fmt.Println("  Bits per label:", cfg.BitsPerLabel)
	fmt.Println("  Units:", opts.NumUnits)
	fmt.Println("  Labels:", uint64(opts.NumUnits)*cfg.LabelsPerUnit)
	fmt.Println("  Size (GiB):", unitSizeInGiB*float32(opts.NumUnits))
	fmt.Println("  Size (Bytes):", unitSizeBytes*uint64(opts.NumUnits))
}

/// setupPos start an interactive proof of space data creation process
func (r *repl) setupPos() {
	cfg, err := r.client.Config()
	if err != nil {
		log.Error("failed get proof of space config from node: %v", err)
		return
	}

	// check if user needs to stop smeshing before changing pos data
	res, err := r.client.IsSmeshing()
	if err != nil {
		log.Error("failed to get proof of space status: %v", err)
		return
	}

	providers, err := r.client.GetPostComputeProviders(false)
	if err != nil {
		log.Error("failed to get compute providers: %v", err)
		return
	}

	if len(providers) == 0 {
		log.Error("No supported compute providers found on system")
		return
	}

	// If smeshing already started, StopSmeshing(false) should be called before init size could be adjusted.
	if res.IsSmeshing {
		stopSmeshing := yesOrNoQuestion("Your node is currently smeshing. To change your proof of space setup, you must first stop smeshing. Would you like to stop smeshing? (y/n)") == "y"
		if stopSmeshing {
			// stop smeshing without deleting the data
			resp, err := r.client.StopSmeshing(false)
			if err != nil {
				log.Error("failed to stop smeshing: %v", err)
				return
			}

			if resp.Code != 0 {
				log.Error("failed to stop smeshing. Response status: %d", resp.Code)
				return
			}

			fmt.Println("Smeshing stopped.")

		} else {
			println("You must stop smeshing before changing your proof of space data setup")
			return
		}
	}

	addrStr := inputNotBlank(enterRewardsAddress)
	addr := gosmtypes.HexToAddress(addrStr)
	dataDir := inputNotBlank(posDataDirMsg)

	if !common.ValidatePath(dataDir) {
		return
	}

	unitSizeBytes := uint64(cfg.BitsPerLabel) * cfg.LabelsPerUnit / 8
	unitSizeInGiB := float32(unitSizeBytes) / float32(gib)
	numUnitsStr := inputNotBlank(fmt.Sprintf(posSizeMsg, unitSizeInGiB, cfg.MinNumUnits, cfg.MaxNumUnits))

	numUnits, err := strconv.ParseUint(numUnitsStr, 10, 32)
	if err != nil {
		log.Error("invalid input: %v", err)
		return
	}

	if uint32(numUnits) > cfg.MaxNumUnits {
		log.Error("Number of units must be equal or less than maximum number of units")
		return
	}

	if uint32(numUnits) < cfg.MinNumUnits {
		log.Error("Number of units must be equal or greater than minimum number of units")
		return
	}

	// validate sufficient free space on target path's volume

	totalSizeBytes := unitSizeBytes * numUnits

	freeSpace, err := common.GetFreeSpace(dataDir)
	if err != nil {
		log.Error("failed to get free space of path's volume: %v", err)
	}

	if totalSizeBytes > freeSpace {
		println("Insufficient free space. Free space: %d, required space: %d", freeSpace, totalSizeBytes)
	}

	// todo: estimate performance for each provider and display performance

	println("Available proof of space compute providers:")
	for _, p := range providers {
		fmt.Printf("Id %d - %s (%s)\n", p.Id, p.Model, computeApiClassName[int32(p.GetComputeApi())])
	}

	providerIdStr := inputNotBlank(posProviderMsg)
	providerId, err := strconv.ParseUint(providerIdStr, 10, 32)
	if err != nil {
		log.Error("invalid input: %v", err)
		return
	}

	validProvider := false
	for _, p := range providers {
		if uint32(providerId) == p.Id {
			validProvider = true
			break
		}
	}

	if !validProvider {
		println("invalid provider id. Please select a provider id for a provider that is available in your system")
		return
	}

	numLabels := numUnits * cfg.LabelsPerUnit
	// request summary information
	fmt.Println("Proof of space setup request summary")
	fmt.Println("Data directory (relative to node or absolute):", dataDir)
	fmt.Println("Size (GiB):", unitSizeInGiB*float32(numUnits))
	fmt.Println("Size (Bytes):", totalSizeBytes)
	fmt.Println("Units:", numUnits)
	fmt.Println("Labels:", numLabels)
	fmt.Println("Bits per label:", cfg.BitsPerLabel)
	fmt.Println("Labels per unit:", cfg.LabelsPerUnit)
	fmt.Println("Compute provider id:", providerId)
	fmt.Println("Data files:", 1)

	req := &apitypes.StartSmeshingRequest{}
	req.Coinbase = &apitypes.AccountId{Address: addr.Bytes()}
	req.Opts = &apitypes.PoSTSetupOpts{
		DataDir:           dataDir,
		NumUnits:          uint32(numUnits),
		NumFiles:          1,
		ComputeProviderId: uint32(providerId),
		Throttle:          false,
	}

	resp, err := r.client.StartSmeshing(req)
	if err != nil {
		log.Error("failed to set up proof of space due to an error: %v", err)
		return
	}

	if resp.Code != 0 {
		log.Error("failed to set up proof of space. Node response code: %d", resp.Code)
		return
	}

	fmt.Println("Proof of space setup has started and your node will start smeshing when it is complete.")
	fmt.Println("IMPORTANT: Please add the following to your node's config file so it will smesh after you restart it.")
	fmt.Println()
	fmt.Println("\"post-init\": {")
	fmt.Printf(" \"datadir\": \"%s\",\n", dataDir)
	fmt.Println(" \"numfiles\": \"1\",")
	fmt.Printf(" \"numunits\": \"%d\",\n", numUnits)
	fmt.Printf(" \"provider\": \"%d\",\n", providerId)
	fmt.Println(" \"throttle\": false,")
	fmt.Println(" \"start-smeshing\": true,")
	fmt.Println("},")
	fmt.Printf("\"coinbase\": \"%s\"\n", addrStr)
	fmt.Println()

	// save pos options in pos.json in cli-wallet directory:
	data, _ := json.MarshalIndent(req.Opts, "", " ")

	err = ioutil.WriteFile(posDataFileName, data, 0644)
	if err == nil {
		fmt.Printf("Saved proof of space setup options to %s.\n\n", posDataFileName)
	} else {
		log.Error("failed to save proof of space setup options to %s: %v", posDataFileName, err)
	}
}

func (r *repl) printPostDataCreationProgress() {
	cfg, err := r.client.Config()
	if err != nil {
		log.Error("failed to query for smeshing config: %v", err)
		return
	}

	stream, err := r.client.PostDataCreationProgressStream()
	if err != nil {
		log.Error("failed to get pos data creation stream: %v", err)
		return
	}

	var initial bool
	for {
		e, err := stream.Recv()
		if err == io.EOF {
			log.Info("api server closed the server-side stream")
			return
		} else if err != nil {
			log.Error("error reading from post data creation stream: %v", err)
			return
		}

		numLabels := uint64(e.Status.Opts.NumUnits) * cfg.LabelsPerUnit
		numLabelsWrittenPct := uint64(float64(e.Status.NumLabelsWritten) / float64(numLabels) * 100)
		PosSizeBytes := uint64(cfg.BitsPerLabel) * numLabels / 8

		if !initial {
			// TODO: use the same printing of cfg/opts as in r.printPostStatus.
			fmt.Printf("Data directory: %s\n", e.Status.Opts.DataDir)
			fmt.Printf("Units: %d\n", e.Status.Opts.NumUnits)
			fmt.Printf("Files: %d\n", e.Status.Opts.NumFiles)
			fmt.Printf("Bits per label: %d\n", cfg.BitsPerLabel)
			fmt.Printf("Labels per unit: %d\n", cfg.LabelsPerUnit)
			fmt.Printf("Labels: %+v\n", numLabels)
			fmt.Printf("Data size: %d bytes\n", PosSizeBytes)
			initial = true
		}

		bytesWritten := uint64(cfg.BitsPerLabel) * e.Status.NumLabelsWritten / 8
		fmt.Printf("Bytes written: %d (%d Labels) - %d%%\n",
			bytesWritten, e.Status.NumLabelsWritten, numLabelsWrittenPct)

		if e.Status.ErrorMessage != "" {
			fmt.Printf("Error: %v\n", e.Status.ErrorMessage)
			return
		}
	}
}

func (r *repl) stopSmeshing() {
	res, err := r.client.IsSmeshing()
	if err != nil {
		log.Error("failed to get proof of space status: %v", err)
		return
	}

	if !res.IsSmeshing {
		fmt.Println("Can't stop smeshing because it has not started")
		return
	}

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

	fmt.Println("Smeshing stopped.\n⚠️  Don't forget to remove smeshing related flags from your node's startup flags (or config file) so it won't start smeshing again after you restart it.")
}

var computeApiClassName = map[int32]string{
	0: "Unspecified",
	1: "CPU",
	2: "CUDA",
	3: "VULKAN",
}

/// setupProofOfSpace prints the available proof of space compute providers
func (r *repl) printPosProviders() {

	providers, err := r.client.GetPostComputeProviders(false)
	if err != nil {
		log.Error("failed to get compute providers: %v", err)
		return
	}

	if len(providers) == 0 {
		fmt.Println("No supported compute providers found")
		return
	}

	fmt.Println("Supported providers on your system:")

	for i, p := range providers {
		if i != 0 {
			fmt.Println("-----")
		}
		fmt.Println("Provider id:", p.GetId())
		fmt.Println("Model:", p.GetModel())
		fmt.Println("Compute api:", computeApiClassName[int32(p.GetComputeApi())])
		// fmt.Println("Performance:", p.GetPerformance())
	}
}

func (r *repl) printSmesherId() {
	if resp, err := r.client.GetSmesherId(); err != nil {
		log.Error("failed to get smesher id: %v", err)
	} else {
		fmt.Println("Smesher id:", "0x"+hex.EncodeToString(resp))
	}
}

func (r *repl) printRewardsAddress() {
	if resp, err := r.client.GetRewardsAddress(); err != nil {
		log.Error("failed to get rewards address: %v", err)
	} else {
		fmt.Println("Rewards address is:", resp.String())
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
		fmt.Println("Rewards address set to:", addr.String())
	} else {
		// todo: what are the possible non-zero status codes here?
		fmt.Printf("Response status code: %d\n", resp.Code)
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

	fmt.Printf("Total rewards: %d\n", total)
	for i, r := range rewards {
		if i != 0 {
			fmt.Println("-----")
		}
		printReward(r)
	}
}

// printSmesherRewards prints all rewards awarded to a smesher identified by an id
func (r *repl) printCurrentSmesherRewards() {
	if smesherId, err := r.client.GetSmesherId(); err != nil {
		log.Error("failed to get smesher id: %v", err)
	} else {

		fmt.Println("Smesher id:", "0x"+hex.EncodeToString(smesherId))

		// todo: request offset and total from user
		rewards, total, err := r.client.SmesherRewards(smesherId, 0, 10000)
		if err != nil {
			log.Error("failed to get rewards: %v", err)
			return
		}

		fmt.Printf("Total rewards: %d\n", total)
		for i, r := range rewards {
			if i != 0 {
				fmt.Println("-----")
			}
			printReward(r)
		}
	}
}
