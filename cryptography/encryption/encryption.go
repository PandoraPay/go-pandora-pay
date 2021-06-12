package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/argon2"
	"io"
)

type EncryptionCipher struct {
	gcm cipher.AEAD
}

func CreateEncryptionCipher(password string, salt []byte) (out *EncryptionCipher, err error) {

	if len(salt) != 32 {
		return nil, errors.New("Salt must be 32 byte")
	}

	key := argon2.IDKey([]byte(password), salt, 100, 32*1024, 4, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	return &EncryptionCipher{
		gcm,
	}, nil

}

func (encryption *EncryptionCipher) Encrypt(data []byte) (out []byte, err error) {

	nonce := make([]byte, encryption.gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	return encryption.gcm.Seal(nonce, nonce, data, nil), nil
}

func (encryption *EncryptionCipher) Decrypt(data []byte) (out []byte, err error) {
	nonceSize := encryption.gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	if out, err = encryption.gcm.Open(nil, nonce, ciphertext, nil); err != nil {
		return nil, err
	}
	return
}
