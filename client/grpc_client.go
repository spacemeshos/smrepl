package client

import (
	"google.golang.org/grpc"
	"strconv"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

const DefaultGRPCPort = 9092
const DefaultGRPCServer = "localhost"

type GRPCClient struct {
	connection *grpc.ClientConn
	server     string
	port       uint
}

func NewGRPCClient(server string, port uint) *GRPCClient {
	return &GRPCClient{
		nil,
		server,
		port,
	}
}

func (c *GRPCClient) Connect() error {
	if c.connection != nil {
		_ = c.connection.Close()
	}

	addr := c.server + ":" + strconv.Itoa(int(c.port))
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

//// services clients

func (c *GRPCClient) nodeServiceClient() pb.NodeServiceClient {
	return pb.NewNodeServiceClient(c.connection)
}

func (c *GRPCClient) debugServiceClient() pb.DebugServiceClient {
	return pb.NewDebugServiceClient(c.connection)
}

func (c *GRPCClient) meshServiceClient() pb.MeshServiceClient {
	return pb.NewMeshServiceClient(c.connection)
}

func (c *GRPCClient) globalStateClient() pb.GlobalStateServiceClient {
	return pb.NewGlobalStateServiceClient(c.connection)
}

func (c *GRPCClient) transactionServiceClient() pb.TransactionServiceClient {
	return pb.NewTransactionServiceClient(c.connection)
}

func (c *GRPCClient) smeshServiceClient() pb.SmesherServiceClient {
	return pb.NewSmesherServiceClient(c.connection)
}

//// Current CLI Wallet commands
