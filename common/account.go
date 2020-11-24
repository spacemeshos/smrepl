package common

import (
	"encoding/hex"
	"fmt"
	"github.com/spacemeshos/ed25519"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
)

type LocalAccount struct {
	Name    string
	PrivKey ed25519.PrivateKey // the pub & private key
	PubKey  ed25519.PublicKey  // only the pub key part
}

func (a *LocalAccount) Address() gosmtypes.Address {
	return gosmtypes.BytesToAddress(a.PubKey[:])
}

type AccountState struct {
	Nonce   uint64
	Balance uint64
}

func (s Store) GetAccount(name string) (*LocalAccount, error) {
	if acc, ok := s[name]; ok {
		priv, err := hex.DecodeString(acc.PrivKey)
		if err != nil {
			return nil, err
		}
		pub, err := hex.DecodeString(acc.PubKey)
		if err != nil {
			return nil, err
		}

		return &LocalAccount{name, priv, pub}, nil
	}
	return nil, fmt.Errorf("account not found")
}

func (s Store) ListAccounts() []string {
	lst := make([]string, 0, len(s))
	for key := range s {
		lst = append(lst, key)
	}
	return lst
}
