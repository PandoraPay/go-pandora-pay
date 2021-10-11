package wallet

import (
	"errors"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography/encryption"
	"pandora-pay/helpers"
)

type WalletEncryption struct {
	wallet           *Wallet
	Encrypted        EncryptedVersion             `json:"encrypted"`
	Salt             []byte                       `json:"salt"`
	Difficulty       int                          `json:"difficulty"`
	password         string                       `json:"-"`
	encryptionCipher *encryption.EncryptionCipher `json:"-"`
}

func createEncryption(wallet *Wallet) *WalletEncryption {
	return &WalletEncryption{
		wallet:    wallet,
		Encrypted: ENCRYPTED_VERSION_PLAIN_TEXT,
	}
}

func (self *WalletEncryption) Encrypt(newPassword string, difficulty int) (err error) {
	self.wallet.Lock()
	defer self.wallet.Unlock()

	if !self.wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	if self.Encrypted != ENCRYPTED_VERSION_PLAIN_TEXT {
		return errors.New("Wallet is encrypted already! Remove the encryption first")
	}

	if difficulty <= 0 || difficulty > 10 {
		return errors.New("Difficulty must be in the interval [1,10]")
	}

	self.Encrypted = ENCRYPTED_VERSION_ENCRYPTION_ARGON2
	self.password = newPassword
	self.Salt = helpers.RandomBytes(32)
	self.Difficulty = difficulty

	if err = self.createEncryptionCipher(); err != nil {
		return
	}

	if err = self.wallet.saveWalletEntire(false); err != nil {
		return
	}

	globals.MainEvents.BroadcastEvent("wallet/encrypted", true)
	return
}

func (self *WalletEncryption) encryptData(input []byte) ([]byte, error) {
	if self.Encrypted == ENCRYPTED_VERSION_ENCRYPTION_ARGON2 {
		return self.encryptionCipher.Encrypt(input)
	}
	return input, nil
}

func (self *WalletEncryption) createEncryptionCipher() (err error) {
	if self.encryptionCipher, err = encryption.CreateEncryptionCipher(self.password, self.Salt, uint32(self.Difficulty)*30); err != nil {
		return
	}
	return
}

func (self *WalletEncryption) Decrypt(password string) (err error) {
	return self.wallet.loadWallet(password, false)
}

func (self *WalletEncryption) decryptData(input []byte) ([]byte, error) {
	if self.Encrypted == ENCRYPTED_VERSION_ENCRYPTION_ARGON2 {
		return self.encryptionCipher.Decrypt(input)
	}
	return input, nil
}

func (self *WalletEncryption) CheckPassword(password string, requirePassword bool) error {
	self.wallet.RLock()
	defer self.wallet.RUnlock()

	if !self.wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	if requirePassword {
		if self.Encrypted == ENCRYPTED_VERSION_PLAIN_TEXT {
			return errors.New("Wallet is not encrypted!")
		}
		if self.password == "" {
			return errors.New("Wallet password was not set!")
		}
	}

	if self.password != password {
		return errors.New("Password is not matching")
	}

	return nil
}

func (self *WalletEncryption) RemoveEncryption() (err error) {
	self.wallet.Lock()
	defer self.wallet.Unlock()

	if !self.wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}
	if self.Encrypted == ENCRYPTED_VERSION_PLAIN_TEXT {
		return errors.New("Wallet is not encrypted!")
	}

	self.Encrypted = ENCRYPTED_VERSION_PLAIN_TEXT
	self.password = ""
	self.Difficulty = 0

	if err = self.wallet.saveWalletEntire(false); err != nil {
		return
	}

	globals.MainEvents.BroadcastEvent("wallet/removed-encryption", true)
	return
}

func (self *WalletEncryption) Logout() (err error) {
	self.wallet.Lock()
	if !self.wallet.Loaded {
		self.wallet.Unlock()
		return
	}
	if self.Encrypted == ENCRYPTED_VERSION_PLAIN_TEXT {
		self.wallet.Unlock()
		return errors.New("Wallet is not encrypted!")
	}
	self.wallet.clearWallet()
	self.wallet.Unlock()

	if err = self.wallet.loadWallet("", true); err != nil {
		return nil
	}

	globals.MainEvents.BroadcastEvent("wallet/logged-out", true)
	return
}
