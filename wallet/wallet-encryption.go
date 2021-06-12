package wallet

import (
	"errors"
	"pandora-pay/cryptography/encryption"
	"pandora-pay/helpers"
)

type WalletEncryption struct {
	wallet           *Wallet
	Encrypted        EncryptedVersion             `json:"encrypted"`
	Salt             []byte                       `json:"encrypted"`
	password         string                       `json:"-"`
	IsOpen           bool                         `json:"-"`
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

	self.wallet.Lock()
	defer self.wallet.Unlock()

	return self.wallet.loadWallet()
}

func (self *WalletEncryption) RemoveEncryption(password string) (err error) {
	self.wallet.Lock()
	defer self.wallet.Unlock()

	self.Encrypted = ENCRYPTED_VERSION_PLAIN_TEXT
	self.password = ""

	return self.wallet.saveWalletEntire(false)
}
