package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spacemeshos/CLIWallet/log"
	"github.com/spacemeshos/ed25519"
)

type AccountKeys struct {
	PubKey  string `json:"pubkey"`
	PrivKey string `json:"privkey"`
}

type Store map[string]AccountKeys

func StoreAccounts(path string, store *Store) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	defer w.Close()
	if err := enc.Encode(store); err != nil {
		return err
	}
	return nil
}

func LoadAccounts(path string) (*Store, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		println("Accounts not loaded from file since it does not exist at: ", path)
		return nil, nil
	}
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening accounts file: %v", err)
	}
	defer r.Close()

	dec := json.NewDecoder(r)
	cfg := &Store{}
	err = dec.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid accounts file content: %v", err)
	}

	return cfg, nil
}

func (s Store) CreateAccount(alias string) *LocalAccount {
	sPub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Error("cannot create account: %s", err)
		return nil
	}
	acc := &LocalAccount{Name: alias, PubKey: sPub, PrivKey: key}
	s[alias] = AccountKeys{PubKey: hex.EncodeToString(sPub), PrivKey: hex.EncodeToString(key)}
	return acc
}
