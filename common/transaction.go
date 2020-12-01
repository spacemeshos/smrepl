package common

import types "github.com/spacemeshos/go-spacemesh/common/types"

// TODO rename to SerializableTransaction once we remove the old SerializableTransaction
type InnerSerializableSignedTransaction struct {
	AccountNonce uint64
	Recipient    types.Address
	GasLimit     uint64
	Price        uint64
	Amount       uint64
}

// Once we support signed txs we should replace SerializableTransaction with this struct. Currently it is only used in the rpc server.
type SerializableSignedTransaction struct {
	InnerSerializableSignedTransaction
	Signature [64]byte
}
