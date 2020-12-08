package repl

import (
	"fmt"
	"time"

	"github.com/spacemeshos/CLIWallet/log"
)

func (r *repl) printMeshInfo() {

	info, err := r.client.GetMeshInfo()
	if err != nil {
		log.Error("failed to get mesh info: %v", err)
		return
	}

	localGenesisTime := time.Unix(int64(info.GenesisTime), 0)

	fmt.Println(printPrefix, "Network Id:", info.NetId)
	fmt.Println(printPrefix, "Max transactions per second:", info.MaxTxsPerSec)
	fmt.Println(printPrefix, "Layers per epoch:", info.LayerPerEpoch)

	fmt.Println(printPrefix, fmt.Sprintf("Layer duration: %d seconds", info.LayerDuration))
	fmt.Println(printPrefix, "Current layer number:", info.CurrentLayer)
	fmt.Println(printPrefix, "Current epoch number:", info.CurrentEpoch)
	fmt.Println(printPrefix, "Genesis time:", localGenesisTime.Local().String())
}
