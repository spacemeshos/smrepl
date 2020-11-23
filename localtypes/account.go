package localtypes

import (
	"encoding/hex"
	"fmt"
	"github.com/spacemeshos/ed25519"
	types "github.com/spacemeshos/go-spacemesh/common/types"
)

type LocalAccount struct {
	Name    string
	PrivKey ed25519.PrivateKey // the pub & private key
	PubKey  ed25519.PublicKey  // only the pub key part
}

func (a *LocalAccount) Address() types.Address {
	return types.BytesToAddress(a.PubKey[:])
}

func StringAddress(addr types.Address) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(addr.Bytes()))
}

func AddressBytesDisplayString(bytes []byte) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(bytes))
}

type AccountInfo struct {
	Nonce   string
	Balance string
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
