package client

import (
	"errors"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"golang.org/x/net/context"
	"strconv"
)

func (c *GRPCClient) NodeURL() string {
	return c.server + ":" + strconv.Itoa(int(c.port)) + " (GRPC API 2.0)"
}

func (c *GRPCClient) Sanity() error {
	service := c.nodeServiceClient()

	const msg = "hello spacemesh"

	resp, err := service.Echo(context.Background(), &pb.EchoRequest{
		Msg: &pb.SimpleString{Value: msg}})

	if err != nil {
		return err
	}

	if resp.Msg.Value != msg {
		return errors.New("unexpected service response")
	}

	return nil
}

func (c *GRPCClient) NodeInfo() (*NodeInfo, error) {
	return nil, nil
}
