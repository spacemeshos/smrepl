package client

import (
	"google.golang.org/grpc"
	"strconv"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
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

func (c *GRPCClient) ServerUrl() string {
	return c.server + ":" + strconv.Itoa(int(c.port)) + " (GRPC API 2.0)"
}

//// services clients

func (c *GRPCClient) nodeServiceClient() apitypes.NodeServiceClient {
	return apitypes.NewNodeServiceClient(c.connection)
}

func (c *GRPCClient) debugServiceClient() apitypes.DebugServiceClient {
	return apitypes.NewDebugServiceClient(c.connection)
}

func (c *GRPCClient) meshServiceClient() apitypes.MeshServiceClient {
	return apitypes.NewMeshServiceClient(c.connection)
}

func (c *GRPCClient) globalStateClient() apitypes.GlobalStateServiceClient {
	return apitypes.NewGlobalStateServiceClient(c.connection)
}

func (c *GRPCClient) transactionServiceClient() apitypes.TransactionServiceClient {
	return apitypes.NewTransactionServiceClient(c.connection)
}

func (c *GRPCClient) smeshServiceClient() apitypes.SmesherServiceClient {
	return apitypes.NewSmesherServiceClient(c.connection)
}

//// Current CLI Wallet commands
