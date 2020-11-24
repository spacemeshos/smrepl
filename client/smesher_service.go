package client

import (
	"context"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *gRPCClient) Smesh(datadir string, space uint, coinbase string) error {
	return nil
}

// SetCoinbase sets the smesher's coinbase address
func (c *gRPCClient) SetCoinbase(address gosmtypes.Address) (*status.Status, error) {
	s := c.smeshServiceClient()

	resp, err := s.SetCoinbase(context.Background(), &apitypes.SetCoinbaseRequest{Id: &apitypes.AccountId{Address: address.Bytes()}})

	if err != nil {
		return nil, err
	}

	return resp.Status, nil

}
