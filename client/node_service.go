package client

import (
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spacemeshos/CLIWallet/common"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"golang.org/x/net/context"
)

// Sanity is a basic api sanity test. It verifies that the client can connect to
// the node service and get a response from it to an echo request.s
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

// NodeInfo returns static node info such as build, version and api server url
func (c *GRPCClient) NodeInfo() (*common.NodeInfo, error) {

	info := &common.NodeInfo{}

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

// NodeStatus returns dynamic node status such as sync status and number of connected peers
func (c *GRPCClient) NodeStatus() (*apitypes.NodeStatus, error) {

	s := c.nodeServiceClient()
	resp, err := s.Status(context.Background(), &apitypes.StatusRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Status, nil
}
