package client

import (
	"context"

	"github.com/spacemeshos/CLIWallet/common"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

func (c *gRPCClient) GlobalStateHash() (*apitypes.GlobalStateHash, error) {

	gsc := c.globalStateClient()

	if resp, err := gsc.GlobalStateHash(context.Background(), &apitypes.GlobalStateHashRequest{}); err != nil {
		return nil, err
	} else {
		return resp.Response, nil
	}
}

// AccountInfo returns basic account data such as balance and nonce from the global state
func (c *gRPCClient) AccountState(address gosmtypes.Address) (*common.AccountState, error) {
	gsc := c.globalStateClient()

	resp, err := gsc.Account(context.Background(), &apitypes.AccountRequest{
		AccountId: &apitypes.AccountId{Address: address.Bytes()}})
	if err != nil {
		return nil, err
	}

	return &common.AccountState{
		Balance: resp.AccountWrapper.StateCurrent.Balance.Value,
		Nonce:   resp.AccountWrapper.StateCurrent.Counter,
	}, nil
}

// SmesherRewards returns rewards for a smesher identified by a smesher id
func (c *gRPCClient) SmesherRewards(smesherId []byte, offset uint32, maxResults uint32) ([]*apitypes.Reward, uint32, error) {
	gsc := c.globalStateClient()

	resp, err := gsc.SmesherDataQuery(context.Background(), &apitypes.SmesherDataQueryRequest{
		SmesherId:  &apitypes.SmesherId{Id: smesherId},
		MaxResults: maxResults,
		Offset:     offset,
	})

	if err != nil {
		return nil, 0, err
	}

	return resp.Rewards, resp.TotalResults, nil
}

func (c *gRPCClient) AccountRewards(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.Reward, uint32, error) {
	gsc := c.globalStateClient()

	resp, err := gsc.AccountDataQuery(context.Background(), &apitypes.AccountDataQueryRequest{
		Filter: &apitypes.AccountDataFilter{
			AccountId:        &apitypes.AccountId{Address: address.Bytes()},
			AccountDataFlags: uint32(apitypes.AccountDataFlag_ACCOUNT_DATA_FLAG_REWARD),
		},

		MaxResults: 0,
		Offset:     0,
	})

	if err != nil {
		return nil, 0, err
	}

	rewards := make([]*apitypes.Reward, 0)

	for _, data := range resp.AccountItem {
		r := data.GetReward()
		// todo: add warning, each result should be a reward
		if r != nil {
			rewards = append(rewards, r)
		}
	}

	return rewards, resp.TotalResults, nil
}

func (c *gRPCClient) AccountTransactionsReceipts(address gosmtypes.Address, offset uint32, maxResults uint32) ([]*apitypes.TransactionReceipt, uint32, error) {
	gsc := c.globalStateClient()

	resp, err := gsc.AccountDataQuery(context.Background(), &apitypes.AccountDataQueryRequest{
		Filter: &apitypes.AccountDataFilter{
			AccountId:        &apitypes.AccountId{Address: address.Bytes()},
			AccountDataFlags: uint32(apitypes.AccountDataFlag_ACCOUNT_DATA_FLAG_TRANSACTION_RECEIPT),
		},

		MaxResults: 0,
		Offset:     0,
	})

	if err != nil {
		return nil, 0, err
	}

	receipts := make([]*apitypes.TransactionReceipt, 0)

	for _, data := range resp.AccountItem {
		r := data.GetReceipt()
		// todo: add warning, each result should be a reward
		if r != nil {
			receipts = append(receipts, r)
		}
	}

	return receipts, resp.TotalResults, nil

}
