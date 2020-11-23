package client

import (
	"context"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *GRPCClient) Smesh(datadir string, space uint, coinbase string) error {
	return nil
}

// SetCoinbase sets the smesher's coinbase address
func (c *GRPCClient) SetCoinbase(address []byte) (*status.Status, error) {
	s := c.smeshServiceClient()

	resp, err := s.SetCoinbase(context.Background(), &apitypes.SetCoinbaseRequest{Id: &apitypes.AccountId{Address: address}})

	if err != nil {
		return nil, err
	}

	return resp.Status, nil

}
