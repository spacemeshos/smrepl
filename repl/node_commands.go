package repl

import (
	"fmt"

	"github.com/spacemeshos/CLIWallet/log"
)

func (r *repl) nodeInfo() {

	info, err := r.client.NodeInfo()
	if err != nil {
		log.Error("failed to get node info: %v", err)
		return
	}

	fmt.Println(printPrefix, "Version:", info.Version)
	fmt.Println(printPrefix, "Build:", info.Build)
	fmt.Println(printPrefix, "API server:", r.client.ServerInfo())

	status, err := r.client.NodeStatus()
	if err != nil {
		log.Error("failed to get node status: %v", err)
		return
	}

	fmt.Println(printPrefix, "Synced:", status.IsSynced)
	fmt.Println(printPrefix, "Synced layer:", status.SyncedLayer.Number)
	fmt.Println(printPrefix, "Current layer:", status.TopLayer.Number)
	fmt.Println(printPrefix, "Verified layer:", status.VerifiedLayer.Number)
	fmt.Println(printPrefix, "Peers:", status.ConnectedPeers)

}
