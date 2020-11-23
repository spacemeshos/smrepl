package client

import (
	"github.com/spacemeshos/CLIWallet/localtypes"
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

//// services

func (c *GRPCClient) nodeServiceClient() pb.NodeServiceClient {
	return pb.NewNodeServiceClient(c.connection)
}

func (c *GRPCClient) debugService() pb.DebugServiceClient {
	return pb.NewDebugServiceClient(c.connection)
}

//// Current CLI Wallet commands

func (c *GRPCClient) AccountInfo(address string) (*localtypes.AccountInfo, error) {
	return nil, nil
}

func (c *GRPCClient) NodeInfo() (*NodeInfo, error) {
	return nil, nil
}

// submit transaction
func (c *GRPCClient) Send(b []byte) (string, error) {
	return "", nil
}

func (c *GRPCClient) Smesh(datadir string, space uint, coinbase string) error {
	return nil
}

func (c *GRPCClient) ListTxs(address string) ([]string, error) {
	txs := make([]string, 0)
	return txs, nil
}

func (c *GRPCClient) SetCoinbase(coinbase string) error {
	return nil
}

func (c *GRPCClient) NodeURL() string {
	return c.server + ":" + strconv.Itoa(int(c.port)) + " (GRPC API 2.0)."
}

func (c *GRPCClient) Sanity() error {
	return nil
}


