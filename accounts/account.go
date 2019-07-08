package accounts

import (
	"encoding/hex"
	"fmt"
	"github.com/spacemeshos/ed25519"
)

type Account struct {
	Name    string
	PrivKey ed25519.PrivateKey // the pub & private key
	PubKey  ed25519.PublicKey  // only the pub key part
}

type AccountInfo struct {
	Nonce   string
	Balance string
}

func (s Store) GetAccount(name string) (*Account, error){
	if acc, ok := s[name]; ok {
		priv, err := hex.DecodeString(acc.PrivKey)
		if err != nil{
			return nil, err
		}
		pub, err := hex.DecodeString(acc.PubKey)
		if err != nil{
			return nil, err
		}

		return &Account{name, priv, pub}, nil
	}
	return nil, fmt.Errorf("account not found")
}

func  (s Store) ListAccounts() []string {
	lst := make([]string, 0, len(s))
	for key := range s {
		lst = append(lst, key)
	}
	return lst
}

