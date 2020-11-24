package client

import (
	"context"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

// GetMeshTransactions returns the transactions on the mesh to or from an address.
func (c *gRPCClient) GetMeshTransactions(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Transaction, uint32, error) {
	ms := c.meshServiceClient()

	resp, err := ms.AccountMeshDataQuery(context.Background(), &apitypes.AccountMeshDataQueryRequest{
		Filter: &apitypes.AccountMeshDataFilter{
			AccountId:            &apitypes.AccountId{Address: address.Bytes()},
			AccountMeshDataFlags: uint32(apitypes.AccountMeshDataFlag_ACCOUNT_MESH_DATA_FLAG_TRANSACTIONS),
		},
		MinLayer:   &apitypes.LayerNumber{Number: 0},
		MaxResults: maxResults,
		Offset:     offset,
	})

	if err != nil {
		return nil, 0, err
	}

	txsMap := make(map[string]bool)
	txs := make([]*apitypes.Transaction, 0)

	for _, data := range resp.Data {
		tx := data.GetTransaction()
		// todo: add warning, each result should be a transaction
		if tx != nil {
			if !txsMap[string(tx.Id.Id)] {
				txsMap[string(tx.Id.Id)] = true
				txs = append(txs, tx)
			}
		}
	}
	// hack alert: for now, we return the number of filtered results and not the results returned from the api
	// because they include duplicated transactions in case where a transaction is on more than 1 mesh block
	return txs, uint32(len(txs)), nil
}

// GetMeshActivations returns activations where the address is the coinbase
func (c *gRPCClient) GetMeshActivations(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Activation, uint32, error) {
	ms := c.meshServiceClient()

	resp, err := ms.AccountMeshDataQuery(context.Background(), &apitypes.AccountMeshDataQueryRequest{
		Filter: &apitypes.AccountMeshDataFilter{
			AccountId:            &apitypes.AccountId{Address: address.Bytes()},
			AccountMeshDataFlags: uint32(apitypes.AccountMeshDataFlag_ACCOUNT_MESH_DATA_FLAG_ACTIVATIONS),
		},
		MinLayer:   &apitypes.LayerNumber{Number: 0},
		MaxResults: maxResults,
		Offset:     offset,
	})

	if err != nil {
		return nil, 0, err
	}

	activations := make([]*apitypes.Activation, 0)

	for _, data := range resp.Data {
		a := data.GetActivation()
		// todo: add warning, each result should be a transaction
		if a != nil {
			activations = append(activations, a)
		}
	}

	return activations, resp.TotalResults, nil
}

type NetInfo struct {
	GenesisTime   uint64
	CurrentLayer  uint32
	CurrentEpoch  uint32
	NetId         uint32
	LayerPerEpoch uint32
	LayerDuration uint32
	MaxTxsPerSec  uint32
}

func (c *gRPCClient) GetNetInfo() (*NetInfo, error) {

	netInfo := &NetInfo{}
	ms := c.meshServiceClient()

	res, err := ms.GenesisTime(context.Background(), &apitypes.GenesisTimeRequest{})
	if err != nil {
		return nil, err
	}

	netInfo.GenesisTime = res.Unixtime.Value

	/*
		res, err := ms.GenesisTime(context.Background(), &apitypes.GenesisTimeRequest{})
		if err != nil {
			return nil, err
		}*/

	return netInfo, nil
}
