package client

import (
	"context"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

// SubmitCoinTransaction submits a signed binary transaction to the node.
func (c *GRPCClient) SubmitCoinTransaction(tx []byte) (*pb.TransactionState, error) {

	s := c.transactionServiceClient()
	resp, err := s.SubmitTransaction(context.Background(), &pb.SubmitTransactionRequest{Transaction: tx})
	if err != nil {
		return nil, err
	}

	return resp.Txstate, nil
}
