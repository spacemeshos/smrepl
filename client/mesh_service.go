package client

import (
	"context"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

func (c *GRPCClient) GetMeshTransactions(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Transaction, uint32, error) {
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

	txs := make([]*apitypes.Transaction, 0)

	for _, data := range resp.Data {
		tx := data.GetTransaction()
		// todo: add warning, each result should be a transaction
		if tx != nil {
			txs = append(txs, tx)
		}
	}

	return txs, resp.TotalResults, nil
}

func (c *GRPCClient) GetMeshActivations(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Activation, uint32, error) {
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
