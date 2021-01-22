package smWallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"time"

	"github.com/spacemeshos/go-spacemesh/common/util"
	"golang.org/x/crypto/pbkdf2"
)

func crypt(block cipher.Block, ciphertext []byte, iv []byte) []byte {
	stream := cipher.NewCTR(block, iv)
	plain := make([]byte, len(ciphertext))
	stream.XORKeyStream(plain, ciphertext)
	return plain
}

func (w *Wallet) twoWayAES(in []byte) ([]byte, error) {
	keyBytes := []byte(w.password)
	key := pbkdf2.Key(keyBytes, []byte(w.Meta.Meta.Salt), 1000000, 32, sha512.New)
	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	iv := make([]byte, c.BlockSize())
	iv[15] = byte(5)
	return crypt(c, in, iv), nil
}

func (w *Wallet) reCrypt() error {
	if len(w.password) == 0 {
		return errors.New("ErrorWalletDoesNotHavePassword")
	}
	privatebuf, err := json.Marshal(w.Crypto.confidential)
	if err != nil {
		return err
	}
	ciphertext, err := w.twoWayAES(privatebuf)
	if err != nil {
		return err
	}
	w.Crypto.CipherText = util.Bytes2Hex(ciphertext)
	if len(w.keystore) > 0 {
		return w.SaveWallet()
	}
	return nil
}

func nowTimeString() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05.000") + "Z"
}
