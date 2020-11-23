package client

import (
	"context"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func (c *GRPCClient) GetMeshTransactions(address []byte, offset uint32, maxResults uint32) ([]*apitypes.Transaction, uint32, error) {
	ms := c.meshServiceClient()

	resp, err := ms.AccountMeshDataQuery(context.Background(), &apitypes.AccountMeshDataQueryRequest{
		Filter: &apitypes.AccountMeshDataFilter{
			AccountId:            &apitypes.AccountId{Address: address},
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
