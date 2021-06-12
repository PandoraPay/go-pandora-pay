package wallet

import (
	"errors"
	"pandora-pay/cryptography/encryption"
	"pandora-pay/helpers"
)

type WalletEncryption struct {
	wallet           *Wallet
	Encrypted        EncryptedVersion             `json:"encrypted"`
	Salt             []byte                       `json:"salt"`
	password         string                       `json:"-"`
	encryptionCipher *encryption.EncryptionCipher `json:"-"`
}

func createEncryption(wallet *Wallet) *WalletEncryption {
	return &WalletEncryption{
		wallet:    wallet,
		Encrypted: ENCRYPTED_VERSION_PLAIN_TEXT,
	}
}

func (self *WalletEncryption) Encrypt(newPassword string) (err error) {
	self.wallet.Lock()
	defer self.wallet.Unlock()

	if self.Encrypted != ENCRYPTED_VERSION_PLAIN_TEXT {
		return errors.New("Wallet is encrypted already! Remove the encryption first")
	}

	self.Encrypted = ENCRYPTED_VERSION_ENCRYPTION_ARGON2
	self.password = newPassword
	self.Salt = helpers.RandomBytes(32)
	if self.encryptionCipher, err = encryption.CreateEncryptionCipher(newPassword, self.Salt); err != nil {
		return
	}

	return self.wallet.saveWalletEntire(false)
}

func (self *WalletEncryption) Decrypt(password string) (err error) {
	return self.wallet.loadWallet(password)
}

func (self *WalletEncryption) RemoveEncryption() (err error) {
	self.wallet.Lock()
	defer self.wallet.Unlock()

	if !self.wallet.loaded {
		return errors.New("Wallet was not loaded!")
	}

	if self.Encrypted == ENCRYPTED_VERSION_PLAIN_TEXT {
		return errors.New("Wallet is not encrypted!")
	}

	self.Encrypted = ENCRYPTED_VERSION_PLAIN_TEXT
	self.password = ""

	return self.wallet.saveWalletEntire(false)
}
