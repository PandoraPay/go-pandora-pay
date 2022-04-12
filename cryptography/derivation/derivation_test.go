package derivation

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
	"testing"
)

func TestDerivePath(t *testing.T) {

	entropy, err := bip39.NewEntropy(256)
	assert.Nil(t, err)

	mnemonic, err := bip39.NewMnemonic(entropy)
	assert.Nil(t, err)

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "SEED Secret Passphrase")
	assert.Nil(t, err)

	assert.Equal(t, len(seed), 64)

	masterKey, err := DeriveForPath(fmt.Sprintf(WebDollarAccountPathFormat, 0), seed)
	assert.Nil(t, err)

	assert.Equal(t, len(masterKey.RawSeed()), 32)

	key2, err := masterKey.Derive(FirstHardenedIndex)
	assert.Nil(t, err)

	assert.Equal(t, len(key2.Key), 32)

}
