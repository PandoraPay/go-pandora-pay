package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/argon2"
	"io"
	"sync"
)

type EncryptionCipher struct {
	gcm cipher.AEAD
	sync.Mutex
}

func CreateEncryptionCipher(password string, salt []byte, time uint32) (out *EncryptionCipher, err error) {

	if len(salt) != 32 {
		return nil, errors.New("Salt must be 32 byte")
	}

	key := argon2.IDKey([]byte(password), salt, time, 32*1024, 4, 32)

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

	return &EncryptionCipher{gcm, sync.Mutex{}}, nil

}

func (encryption *EncryptionCipher) Encrypt(data []byte) (out []byte, err error) {

	encryption.Lock()
	defer encryption.Unlock()

	nonce := make([]byte, encryption.gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	return encryption.gcm.Seal(nonce, nonce, data, nil), nil
}

func (encryption *EncryptionCipher) Decrypt(data []byte) (out []byte, err error) {

	encryption.Lock()
	defer encryption.Unlock()

	nonceSize := encryption.gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	if out, err = encryption.gcm.Open(nil, nonce, ciphertext, nil); err != nil {
		return nil, err
	}
	return
}
