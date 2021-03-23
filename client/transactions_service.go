package client

import (
	"context"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

// SubmitCoinTransaction submits a signed binary transaction to the node.
func (c *gRPCClient) SubmitCoinTransaction(tx []byte) (*apitypes.TransactionState, error) {
	s := c.getTransactionServiceClient()
	resp, err := s.SubmitTransaction(context.Background(), &apitypes.SubmitTransactionRequest{Transaction: tx})
	if err != nil {
		return nil, err
	}

	return resp.Txstate, nil
}

// TransactionState returns the state and optionally the transaction for a single transaction based on tx id
func (c *gRPCClient) TransactionState(txId []byte, includeTx bool) (*apitypes.TransactionState, *apitypes.Transaction, error) {
	s := c.getTransactionServiceClient()
	ids := make([]*apitypes.TransactionId, 0)
	ids = append(ids, &apitypes.TransactionId{Id: txId})

	resp, err := s.TransactionsState(context.Background(), &apitypes.TransactionsStateRequest{
		TransactionId:       ids,
		IncludeTransactions: includeTx,
	})

	if err != nil {
		return nil, nil, err
	}
	return resp.TransactionsState[0], resp.Transactions[0], nil
}
