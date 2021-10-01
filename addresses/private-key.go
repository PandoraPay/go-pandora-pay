package addresses

import (
	"context"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key helpers.HexBytes `json:"key"` //32 byte
}

func (pk *PrivateKey) GeneratePublicKey() []byte {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	publicKey := crypto.GPoint.ScalarMult(priv)
	return publicKey.EncodeCompressed()
}

func (pk *PrivateKey) GenerateAddress(registration bool, amount uint64, paymentID []byte) (*Address, error) {
	publicKey := pk.GeneratePublicKey()

	var reg []byte
	var err error
	if registration {
		if reg, err = pk.GetRegistration(); err != nil {
			return nil, err
		}
	}

	return NewAddr(config.NETWORK_SELECTED, SIMPLE_PUBLIC_KEY, publicKey, reg, amount, paymentID)
}

func (pk *PrivateKey) GetRegistration() ([]byte, error) {
	return pk.Sign([]byte("registration"))
}

//make sure message is a hash to avoid leaking any parts of the private key
func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	return crypto.SignMessage(message, pk.Key)
}

func (pk *PrivateKey) Decrypt(message []byte) ([]byte, error) {
	panic("not implemented")
}

func (pk *PrivateKey) DecodeBalance(balance *crypto.ElGamal, previousValue uint64, ctx context.Context) (uint64, error) {
	priv := new(crypto.BNRed).SetBytes(pk.Key)
	balancePoint := new(bn256.G1).Add(balance.Left, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(balance.Right, priv.BigInt())))
	return crypto.BalanceDecoder.BalanceDecode(balancePoint, previousValue, ctx)
}

func GenerateNewPrivateKey() *PrivateKey {
	seed := crypto.RandomScalarBNRed()
	privateKey := seed.ToBytes()

	return &PrivateKey{Key: privateKey}
}

func CreatePrivateKeyFromSeed(seed []byte) *PrivateKey {
	return &PrivateKey{Key: seed}
}
