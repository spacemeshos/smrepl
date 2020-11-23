package client

import (
	"google.golang.org/grpc"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
	"strconv"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

type GRPCClient struct {
	connection *grpc.ClientConn
	server string
	port uint
}

func NewGRPCClient(server string, port uint) *GRPCClient {
	return &GRPCClient {
		nil,
		server,
		port,
	}
}

func (c *GRPCClient) Connect() error {
	addr := c.server + ":" + strconv.Itoa(int(c.port))
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.connection = conn
	return nil
}

func (c *GRPCClient) Close() error {
	 if c.connection != nil {
		 return c.connection.Close()
	 }
	 return nil
}

func (c *GRPCClient) NodeServiceClient() pb.NodeServiceClient {
	return pb.NewNodeServiceClient(c.connection)
}

func (c *GRPCClient) DebugService() pb.DebugServiceClient {
	return pb.NewDebugServiceClient(c.connection)
}

