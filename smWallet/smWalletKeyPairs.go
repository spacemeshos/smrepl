package smWallet

import (
	"encoding/hex"
	hx "encoding/hex"
	"errors"
	"fmt"

	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/tyler-smith/go-bip39"
)

/*
PublicKey retrieves public key from private
*/
func PublicKey(key ed25519.PrivateKey) ed25519.PublicKey {
	return key.Public().(ed25519.PublicKey)
}

func (w *Wallet) newAccount(displayName string) (*account, error) {
	if !w.unlocked {
		return nil, errors.New(errorWalletNotUnlocked)
	}
	seed := bip39.NewSeed(w.Crypto.confidential.Mnemonic, "")
	i := uint64(0)
	for {
		pk := ed25519.NewDerivedKeyFromSeed(seed[:32], i, []byte(spaceSalt))
		pub := pk.Public().(ed25519.PublicKey)[:]
		addr := types.BytesToAddress(pub)
		found := false
		for _, acc := range w.Crypto.confidential.Accounts {
			if addr == acc.Address() {
				found = true
				break
			}
		}
		if !found {
			ac := account{
				DisplayName: displayName,
				Created:     nowTimeString(),
				Path:        "",
				PublicKey:   hx.EncodeToString(pub),
				SecretKey:   hx.EncodeToString(pk),
			}
			return &ac, nil
		}
		i++
	}
}

// GenerateNewPair - add a new pair based on mnemonic key phrase
func (w *Wallet) GenerateNewPair(displayName string) (int, error) {
	ac, err := w.newAccount(displayName)
	if err != nil {
		return 0, err
	}
	w.Crypto.confidential.Accounts = append(w.Crypto.confidential.Accounts, *ac)
	err = w.reCrypt()
	if err != nil {
		return 0, err
	}
	return len(w.Crypto.confidential.Accounts) - 1, nil
}

func (w *Wallet) verifyAccounts() (err error) {
	message := []byte{5, 4, 3, 2, 1}
	for pos, acc := range w.Crypto.confidential.Accounts {
		var secret ed25519.PrivateKey
		var public ed25519.PublicKey
		secret, err = hex.DecodeString(acc.SecretKey)
		if err != nil {
			return err
		}

		public, err := hex.DecodeString(acc.PublicKey)
		if err != nil {
			return err
		}
		sig := ed25519.Sign(secret, message)
		ok := ed25519.Verify(public, message, sig)
		if !ok {
			return fmt.Errorf("error verifying account %d", pos)
		}
	}
	return nil
}
