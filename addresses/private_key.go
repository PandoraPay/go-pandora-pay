package addresses

import (
	"context"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/cryptography/crypto/balance-decryptor"
)

type PrivateKey struct {
	Key []byte `json:"key" msgpack:"key"` //32 byte
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

	version := SIMPLE_PUBLIC_KEY

	var reg []byte
	var err error

	if registration {
		if reg, err = pk.GetRegistration(staked, spendPublicKey); err != nil {
			return nil, err
		}
	}

	return NewAddr(config.NETWORK_SELECTED, version, publicKey, staked, spendPublicKey, reg, paymentID, paymentAmount, paymentAsset)
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

func (pk *PrivateKey) DecryptBalance(balance *crypto.ElGamal, previousValue uint64, ctx context.Context, statusCallback func(string)) (uint64, error) {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	balancePoint := new(bn256.G1).Add(balance.Left, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(balance.Right, priv.BigInt())))
	return balance_decryptor.BalanceDecryptor.DecryptBalance(balancePoint, previousValue, ctx, statusCallback)
}

func (pk *PrivateKey) TryDecryptBalance(balance *crypto.ElGamal, matchValue uint64) bool {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	balancePoint := new(bn256.G1).Add(balance.Left, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(balance.Right, priv.BigInt())))
	return balance_decryptor.BalanceDecryptor.TryDecryptBalance(balancePoint, matchValue)
}

func GenerateNewPrivateKey() *PrivateKey {
	seed := crypto.RandomScalarBNRed()
	privateKey := seed.ToBytes()

	return &PrivateKey{Key: privateKey}
}

func CreatePrivateKeyFromSeed(key []byte) (*PrivateKey, error) {
	if len(key) != cryptography.PrivateKeySize {
		return nil, errors.New("Private key length is invalid")
	}
	return &PrivateKey{Key: key}, nil
}
