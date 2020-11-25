package client

import (
	"strconv"

	"google.golang.org/grpc"

	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

const DefaultGRPCPort = 9092
const DefaultGRPCServer = "localhost"

type gRPCClient struct {
	connection               *grpc.ClientConn
	server                   string
	port                     uint
	nodeServiceClient        apitypes.NodeServiceClient
	debugServiceClient       apitypes.DebugServiceClient
	meshServiceClient        apitypes.MeshServiceClient
	globalStateServiceClient apitypes.GlobalStateServiceClient
	transactionServiceClient apitypes.TransactionServiceClient
	smesherServiceClient     apitypes.SmesherServiceClient
}

func newGRPCClient(server string, port uint) *gRPCClient {
	return &gRPCClient{
		nil,
		server,
		port,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	}
}

func (c *gRPCClient) Connect() error {
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

func (c *gRPCClient) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

func (c *gRPCClient) ServerUrl() string {
	return c.server + ":" + strconv.Itoa(int(c.port)) + " (GRPC API 2.0)"
}

//// services clients

func (c *gRPCClient) getNodeServiceClient() apitypes.NodeServiceClient {
	if c.nodeServiceClient == nil {
		c.nodeServiceClient = apitypes.NewNodeServiceClient(c.connection)
	}
	return c.nodeServiceClient
}

func (c *gRPCClient) getDebugServiceClient() apitypes.DebugServiceClient {
	if c.debugServiceClient == nil {
		c.debugServiceClient = apitypes.NewDebugServiceClient(c.connection)
	}
	return c.debugServiceClient
}

func (c *gRPCClient) getMeshServiceClient() apitypes.MeshServiceClient {
	if c.meshServiceClient == nil {
		c.meshServiceClient = apitypes.NewMeshServiceClient(c.connection)
	}
	return c.meshServiceClient
}

func (c *gRPCClient) getGlobalStateServiceClient() apitypes.GlobalStateServiceClient {
	if c.globalStateServiceClient == nil {
		c.globalStateServiceClient = apitypes.NewGlobalStateServiceClient(c.connection)
	}
	return c.globalStateServiceClient

}

func (c *gRPCClient) getTransactionServiceClient() apitypes.TransactionServiceClient {
	if c.transactionServiceClient == nil {
		c.transactionServiceClient = apitypes.NewTransactionServiceClient(c.connection)
	}
	return c.transactionServiceClient
}

func (c *gRPCClient) getSmesherServiceClient() apitypes.SmesherServiceClient {
	if c.smesherServiceClient == nil {
		c.smesherServiceClient = apitypes.NewSmesherServiceClient(c.connection)
	}
	return c.smesherServiceClient
}

//// Current CLI Wallet commands
