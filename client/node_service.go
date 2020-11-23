package client

import (
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spacemeshos/CLIWallet/localtypes"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"golang.org/x/net/context"
)

func (c *GRPCClient) Sanity() error {
	service := c.nodeServiceClient()

	const msg = "hello spacemesh"

	resp, err := service.Echo(context.Background(), &apitypes.EchoRequest{
		Msg: &apitypes.SimpleString{Value: msg}})

	if err != nil {
		return err
	}

	if resp.Msg.Value != msg {
		return errors.New("unexpected node service echo response")
	}

	return nil
}

func (c *GRPCClient) NodeInfo() (*localtypes.NodeInfo, error) {

	info := &localtypes.NodeInfo{}

	s := c.nodeServiceClient()
	resp, err := s.Version(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}
	info.Version = resp.VersionString.Value

	resp1, err := s.Build(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}
	info.Build = resp1.BuildString.Value

	return info, nil
}

func (c *GRPCClient) NodeStatus() (*apitypes.NodeStatus, error) {

	s := c.nodeServiceClient()
	resp, err := s.Status(context.Background(), &apitypes.StatusRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Status, nil
}
