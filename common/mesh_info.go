package common

type NetInfo struct {
	GenesisTime   uint64
	CurrentLayer  uint32
	CurrentEpoch  uint64
	NetId         uint64
	LayerPerEpoch uint64
	LayerDuration uint64
	MaxTxsPerSec  uint64
}
