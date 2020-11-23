package client

import (
	"context"
	"github.com/spacemeshos/CLIWallet/localtypes"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func (c *GRPCClient) AccountInfo(address []byte) (*localtypes.AccountState, error) {
	gsc := c.globalStateClient()

	resp, err := gsc.Account(context.Background(), &pb.AccountRequest{
		AccountId: &pb.AccountId{Address: address}})
	if err != nil {
		return nil, err
	}

	return &localtypes.AccountState{
		Balance: resp.AccountWrapper.StateCurrent.Balance.Value,
		Nonce:   resp.AccountWrapper.StateCurrent.Counter,
	}, nil

}
