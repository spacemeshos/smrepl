package client

import (
	"context"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func (c *GRPCClient) GetMeshTransactions(address []byte, offset uint32, maxResults uint32) ([]*pb.Transaction, uint32, error) {
	ms := c.meshServiceClient()

	resp, err := ms.AccountMeshDataQuery(context.Background(), &pb.AccountMeshDataQueryRequest{
		Filter: &pb.AccountMeshDataFilter{
			AccountId:            &pb.AccountId{Address: address},
			AccountMeshDataFlags: uint32(pb.AccountMeshDataFlag_ACCOUNT_MESH_DATA_FLAG_TRANSACTIONS),
		},
		MinLayer:   &pb.LayerNumber{Number: 0},
		MaxResults: maxResults,
		Offset:     offset,
	})

	if err != nil {
		return nil, 0, err
	}

	txs := make([]*pb.Transaction, 0)

	for _, data := range resp.Data {
		tx := data.GetTransaction()
		// todo: add warning, each result should be a transaction
		if tx != nil {
			txs = append(txs, tx)
		}
	}

	return txs, resp.TotalResults, nil
}
