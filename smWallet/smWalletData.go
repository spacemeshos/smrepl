package smWallet

import (
	"errors"

	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/common/types"
)

// GetMnemonic returns the mnemonic string associated with the wallet
func (w Wallet) GetMnemonic() (string, error) {
	if !w.unlocked {
		return "", errors.New(ErrorWalletNotUnlocked)
	}
	return w.Crypto.confidential.Mnemonic, nil
}

// GetNumberOfAccounts returns the number of accounts held in said wallet
func (w Wallet) GetNumberOfAccounts() (int, error) {
	if !w.unlocked {
		return 0, errors.New(ErrorWalletNotUnlocked)
	}
	return len(w.Crypto.confidential.Accounts), nil
}

// GetAddress retrieves an address from a wallet if unlocked and it has been generated
func (w *Wallet) GetAddress(accountNumber int) (types.Address, error) {
	if !w.unlocked {
		return types.Address{}, errors.New(ErrorWalletNotUnlocked)
	}
	if accountNumber >= len(w.Crypto.confidential.Accounts) {
		return types.Address{}, errors.New(ErrorWalletDoesNotHaveThatAddress)
	}
	return w.Crypto.confidential.Accounts[accountNumber].Address(), nil
}

// GetPublicKey retrieves an address from a wallet if unlocked and it has been generated
func (w *Wallet) GetPublicKey(accountNumber int) (ed25519.PublicKey, error) {
	private, err := w.GetPrivateKey(accountNumber)
	if err != nil {
		return []byte{}, err
	}
	return PublicKey(private), nil
}

// GetPrivateKey retrieve the private key
func (w *Wallet) GetPrivateKey(accountNumber int) (ed25519.PrivateKey, error) {
	if !w.unlocked {
		return []byte{}, errors.New(ErrorWalletNotUnlocked)
	}
	if accountNumber >= len(w.Crypto.confidential.Accounts) {
		return []byte{}, errors.New(ErrorWalletDoesNotHaveThatAddress)
	}
	thisAccount := w.Crypto.confidential.Accounts[accountNumber]
	private, err := thisAccount.PrivateKey()
	if err != nil {
		return []byte{}, err
	}
	return private, nil
}

// GetAccountDisplayName retrieves an account name from a wallet (if unlocked and account exists)
func (w *Wallet) GetAccountDisplayName(accountNumber int) (string, error) {
	if !w.unlocked {
		return "", errors.New(ErrorWalletNotUnlocked)
	}
	if accountNumber >= len(w.Crypto.confidential.Accounts) {
		return "", errors.New(ErrorWalletDoesNotHaveThatAddress)
	}
	return w.Crypto.confidential.Accounts[accountNumber].DisplayName, nil
}

// SetCurrent - set current wallet by number
func (w *Wallet) SetCurrent(accountNumber int) error {
	if !w.unlocked {
		return errors.New(ErrorWalletNotUnlocked)
	}
	if accountNumber >= len(w.Crypto.confidential.Accounts) {
		return errors.New(ErrorWalletDoesNotHaveThatAddress)
	}
	w.Crypto.confidential.accountNumber = accountNumber
	return nil
}

func (w *Wallet) AddContact(nickname string, address types.Address) error {
	if !w.unlocked {
		return errors.New(ErrorWalletNotUnlocked)
	}
	// TO DO : Validate nickname and address
	w.Crypto.confidential.Contacts = append(w.Crypto.confidential.Contacts, contact{Nickname: nickname, Address: address.Hex()})
	err := w.reCrypt()
	return err
}
