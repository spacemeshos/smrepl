package common

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/ed25519"
)

type accountKeys struct {
	PubKey  string `json:"pubkey"`
	PrivKey string `json:"privkey"`
}

type Store map[string]accountKeys

func (s Store) CreateAccount(alias string) *LocalAccount {
	sPub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Error("cannot create account: %s", err)
		return nil
	}
	acc := &LocalAccount{Name: alias, PubKey: sPub, PrivKey: key}
	s[alias] = accountKeys{PubKey: hex.EncodeToString(sPub), PrivKey: hex.EncodeToString(key)}
	return acc
}
