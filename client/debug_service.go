package client

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func (c *GRPCClient) DebugAllAccounts() ([]*pb.Account, error) {
	dbgService := c.debugServiceClient()
	resp, err := dbgService.Accounts(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return resp.AccountWrapper, nil

}
