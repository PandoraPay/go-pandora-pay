package addresses

import (
	"context"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/cryptography/crypto/balance_decryptor"
)

type PrivateKey struct {
	KeyWIF
}

func (pk *PrivateKey) GeneratePublicKeyPoint() *bn256.G1 {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	publicKey := crypto.GPoint.ScalarMult(priv)
	return publicKey.G1()
}

func (pk *PrivateKey) GeneratePublicKey() []byte {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	publicKey := crypto.GPoint.ScalarMult(priv)
	return publicKey.EncodeCompressed()
}

func (pk *PrivateKey) GenerateAddress(staked bool, spendPublicKey []byte, registration bool, paymentID []byte, paymentAmount uint64, paymentAsset []byte) (*Address, error) {
	publicKey := pk.GeneratePublicKey()

	var reg []byte
	var err error

	if registration {
		if reg, err = pk.GetRegistration(staked, spendPublicKey); err != nil {
			return nil, err
		}
	}

	return CreateAddr(publicKey, staked, spendPublicKey, reg, paymentID, paymentAmount, paymentAsset)
}

func (pk *PrivateKey) GetRegistration(staked bool, spendPublicKey []byte) ([]byte, error) {
	data := []byte("registration")
	if staked {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	data = append(data, spendPublicKey...)
	return pk.Sign(data)
}

//make sure message is a hash to avoid leaking any parts of the private key
func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	return crypto.SignMessage(message, pk.Key)
}

func (pk *PrivateKey) Decrypt(message []byte) ([]byte, error) {
	return nil, errors.New("Encryption is not supported right now")
}

func (pk *PrivateKey) DecryptBalance(balance *crypto.ElGamal, tryPreviousValue bool, previousValue uint64, ctx context.Context, statusCallback func(string)) (uint64, error) {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	balancePoint := new(bn256.G1).Add(balance.Left, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(balance.Right, priv.BigInt())))
	return balance_decryptor.BalanceDecryptor.DecryptBalance(balancePoint, tryPreviousValue, previousValue, ctx, statusCallback)
}

func (pk *PrivateKey) TryDecryptBalance(balance *crypto.ElGamal, matchValue uint64) bool {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	balancePoint := new(bn256.G1).Add(balance.Left, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(balance.Right, priv.BigInt())))
	return balance_decryptor.BalanceDecryptor.TryDecryptBalance(balancePoint, matchValue)
}

func (pk *PrivateKey) Deserialize(buffer []byte) error {
	return pk.deserialize(buffer, cryptography.PrivateKeySize)
}

func GenerateNewPrivateKey() *PrivateKey {
	for {
		seed := crypto.RandomScalarBNRed()
		key := seed.ToBytes()

		privateKey, err := NewPrivateKey(key)
		if err != nil {
			continue
		}
		return privateKey
	}
}

func NewPrivateKey(key []byte) (*PrivateKey, error) {

	if len(key) != cryptography.PrivateKeySize {
		return nil, errors.New("Private Key length is invalid")
	}

	privateKey := &PrivateKey{
		KeyWIF{
			SIMPLE_PRIVATE_KEY_WIF,
			config.NETWORK_SELECTED,
			key,
			nil,
		},
	}

	privateKey.Checksum = privateKey.computeCheckSum()

	return privateKey, nil
}
