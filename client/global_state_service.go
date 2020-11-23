package client

import (
	"context"
	"github.com/spacemeshos/CLIWallet/localtypes"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func (c *GRPCClient) AccountInfo(address []byte) (*localtypes.AccountState, error) {
	gss := c.globalStateClient()

	resp, err := gss.Account(context.Background(), &pb.AccountRequest{
		AccountId: &pb.AccountId{Address: address}})
	if err != nil {
		return nil, err
	}

	return &localtypes.AccountState{
		Balance: resp.AccountWrapper.StateCurrent.Balance.Value,
		Nonce:   resp.AccountWrapper.StateCurrent.Counter,
	}, nil

}

func (c *GRPCClient) ListTxs(address string) ([]string, error) {
	//gss := c.globalStateClient()
	txs := make([]string, 0)
	return txs, nil
}
