package client

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// GetSmesherId returns the current smesher id configured in the node
func (c *gRPCClient) GetSmesherId() ([]byte, error) {
	s := c.smeshServiceClient()
	if resp, err := s.SmesherID(context.Background(), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		return resp.AccountId.Address, nil
	}
}

// IsSmeshing returns true iff the node is currently setup to smesh
func (c *gRPCClient) IsSmeshing() (bool, error) {
	s := c.smeshServiceClient()
	if resp, err := s.IsSmeshing(context.Background(), &empty.Empty{}); err != nil {
		return false, err
	} else {
		return resp.IsSmeshing, nil
	}
}

// StartSmeshing instructs the node to start smeshing using user's provider params
func (c *gRPCClient) StartSmeshing(address gosmtypes.Address, dataDir string, dataSizeBytes uint64) (*status.Status, error) {
	s := c.smeshServiceClient()

	resp, err := s.StartSmeshing(context.Background(), &apitypes.StartSmeshingRequest{
		Coinbase:       &apitypes.AccountId{Address: address.Bytes()},
		DataDir:        dataDir,
		CommitmentSize: &apitypes.SimpleInt{Value: dataSizeBytes},
	})

	if err != nil {
		return nil, err
	}

	return resp.Status, nil
}

// StopSmeshing instructs the node to stop smeshing and optionally delete smeshing data file(s)
func (c *gRPCClient) StopSmeshing(deleteFiles bool) (*status.Status, error) {
	s := c.smeshServiceClient()

	resp, err := s.StopSmeshing(context.Background(), &apitypes.StopSmeshingRequest{DeleteFiles: deleteFiles})

	if err != nil {
		return nil, err
	}

	return resp.Status, nil
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

// GetCoinbase get the smesher's current rewards address
func (c *gRPCClient) GetCoinbase() (*gosmtypes.Address, error) {
	s := c.smeshServiceClient()

	resp, err := s.Coinbase(context.Background(), &empty.Empty{})

	if err != nil {
		return nil, err
	}

	addr := gosmtypes.BytesToAddress(resp.AccountId.Address)

	return &addr, nil

}
