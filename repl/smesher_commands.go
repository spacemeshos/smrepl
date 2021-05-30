package repl

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"

	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/go-spacemesh/common/util"
)

// GIB is the number of bytes in 1 GiByes
const GIB uint64 = 1_262_485_504
const pos_data_file_name = "pos-data.json"

func (r *repl) printSmeshingStatus() {
	res, err := r.client.SmeshingStatus()
	if err != nil {
		log.Error("failed to get proof of space status: %v", err)
		return
	}

	switch res.Status {
	case apitypes.SmeshingStatusResponse_SMESHING_STATUS_IDLE:
		fmt.Println(printPrefix, "Proof of space data was not created.")
	case apitypes.SmeshingStatusResponse_SMESHING_STATUS_CREATING_POST_DATA:
		fmt.Println(printPrefix, "⏱ Proof of space data creation is in progress.")
	case apitypes.SmeshingStatusResponse_SMESHING_STATUS_ACTIVE:
		fmt.Println(printPrefix, "👍 Proof of space data was created and is used for smeshing.")
	default:
		fmt.Println("printPrefix", "Unexpected api result.")
	}

	cfg, err := r.client.Config()
	if err != nil {
		log.Error("failed get proof of space config from node: %v", err)
		return
	}

	data, err := ioutil.ReadFile(pos_data_file_name)
	if err != nil {
		log.Error("failed to read proof of space from data file: %v", err)
		fmt.Println(printPrefix, "failed to read from %s. Error: %v", pos_data_file_name, err)
	} else {
		var posInitOps apitypes.PostInitOpts
		err = json.Unmarshal(data, &posInitOps)
		if err != nil {
			log.Error("failed to parse data from %s. %v", pos_data_file_name, err)
			fmt.Println(printPrefix, "failed to parse data from %s. Error: %v", pos_data_file_name, err)
			return
		}

		fmt.Println("Proof of space information:")

		fmt.Println("Data dir (relative to node):", posInitOps.DataDir)
		fmt.Println("Date files:", posInitOps.NumFiles)
		fmt.Println("Compute provider id:", posInitOps.ComputeProviderId)
		fmt.Println("Throttle when computer is busy:", posInitOps.Throttle)
		fmt.Println("Bits per label:", cfg.BitsPerLabel)

		fmt.Println("Units:", posInitOps.NumUnits)
		fmt.Println("Labels:", uint64(posInitOps.NumUnits)*cfg.LabelsPerUnit)

		unitSizeBytes := uint64(cfg.BitsPerLabel) * cfg.LabelsPerUnit / 8
		unitSizeInGiB := float32(unitSizeBytes) / float32(GIB)

		fmt.Println("Size (GiB):", unitSizeInGiB*float32(posInitOps.NumUnits))
		fmt.Println("Size (Bytes):", unitSizeBytes*uint64(posInitOps.NumUnits))

	}
}

/// setupPos start an interactive proof of space data creation process
func (r *repl) setupPos() {
	cfg, err := r.client.Config()
	if err != nil {
		log.Error("failed get proof of space config from node: %v", err)
		return
	}

	// check if user needs to stop smeshing before changing pos data
	res, err := r.client.SmeshingStatus()
	if err != nil {
		log.Error("failed to get proof of space status: %v", err)
		return
	}

	// check if idle - if not idle then pos is in progress or pos is active....
	// in both cases we need to call StopSmeshing(false) before pos size can be adjusted
	if res.Status != apitypes.SmeshingStatusResponse_SMESHING_STATUS_IDLE {
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

			fmt.Println(printPrefix, "Smeshing stopped.")

		} else {
			println("You must stop smeshing before changing your proof of space data setup")
			return
		}
	}

	addrStr := inputNotBlank(enterRewardsAddress)
	addr := gosmtypes.HexToAddress(addrStr)
	dataDir := inputNotBlank(posDataDirMsg)

	unitSizeBytes := uint64(cfg.BitsPerLabel) * cfg.LabelsPerUnit / 8
	unitSizeInGiB := float32(unitSizeBytes) / float32(GIB)
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

	// TODO: validate provider id is valid by enum the providers here....

	providerIdStr := inputNotBlank(posProviderMsg)
	providerId, err := strconv.ParseUint(providerIdStr, 10, 32)
	if err != nil {
		log.Error("invalid input: %v", err)
		return
	}

	totalSizeBytes := unitSizeBytes * numUnits
	numLabels := numUnits * cfg.LabelsPerUnit
	// request summary information
	fmt.Println(printPrefix, "Proof of space setup request summary")
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
	req.Opts = &apitypes.PostInitOpts{
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

	fmt.Println(printPrefix, "Proof of space setup has started and your node will start smeshing when it is complete.")
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

	// save pos options in pos.json in cliwallet folder:
	data, _ := json.MarshalIndent(req.Opts, "", " ")

	err = ioutil.WriteFile(pos_data_file_name, data, 0644)
	if err == nil {
		fmt.Printf("Saved proof of space setup options to %s.\n\n", pos_data_file_name)
	} else {
		log.Error("failed to save proof of space setup options to %s: %v", pos_data_file_name, err)
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

		numLabels := uint64(e.Status.SessionOpts.NumUnits) * cfg.LabelsPerUnit
		numLabelsWrittenPct := uint64(float64(e.Status.NumLabelsWritten) / float64(numLabels) * 100)
		PosSizeBytes := uint64(cfg.BitsPerLabel) * numLabels / 8

		if initial == false {
			fmt.Printf("Data directory: %s\n", e.Status.SessionOpts.DataDir)
			fmt.Printf("Units: %d\n", e.Status.SessionOpts.NumUnits)
			fmt.Printf("Files: %d\n", e.Status.SessionOpts.NumFiles)
			fmt.Printf("Bits per label: %d\n", cfg.BitsPerLabel)
			fmt.Printf("Labels per unit: %d\n", cfg.LabelsPerUnit)
			fmt.Printf("Labels: %+v\n", numLabels)
			fmt.Printf("Data size: %d bytes\n", PosSizeBytes)
			initial = true
		}

		bytesWritten := uint64(cfg.BitsPerLabel) * e.Status.NumLabelsWritten / 8

		fmt.Printf("Bytes written: %d (%d Labels) - %d%%\n",
			bytesWritten, e.Status.NumLabelsWritten, numLabelsWrittenPct)
	}
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

	fmt.Println(printPrefix, "Smeshing stopped.\n⚠️  Don't forget to remove smeshing related flags from your node's startup flags (or config file) so it won't start smeshing again after you restart it.")
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
		fmt.Println(printPrefix, "No supported compute providers found")
		return
	}

	fmt.Println(printPrefix, "Supported providers on your system:")

	for i, p := range providers {
		if i != 0 {
			fmt.Println("-----")
		}
		fmt.Println("Provider id:", p.GetId())
		fmt.Println("Model:", p.GetModel())
		fmt.Println("Compute api:", computeApiClassName[int32(p.GetComputeApi())])
		fmt.Println("Performance:", p.GetPerformance())
	}
}

func (r *repl) print() {
	providers, err := r.client.GetPostComputeProviders(false)
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
			fmt.Println("-----")
		}
		fmt.Println("Provider id:", p.GetId())
		fmt.Println("Model:", p.GetModel())
		fmt.Println("Compute api:", computeApiClassName[int32(p.GetComputeApi())])
		fmt.Println("Performance:", p.GetPerformance())
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
				fmt.Println("-----")
			}
			printReward(r)
		}
	}
}
