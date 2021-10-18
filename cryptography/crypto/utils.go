package crypto

import (
	"bytes"
	"fmt"
	"math/big"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
)

func SignMessage(message, key []byte) ([]byte, error) {

	var tmppoint bn256.G1
	tmpsecret := RandomScalar()
	tmppoint.ScalarMult(G, tmpsecret)

	priv := new(BNRed).SetBytes(key)
	pubKey := GPoint.ScalarMult(priv)

	serialize := []byte(fmt.Sprintf("%s%s%s", pubKey.G1().String(), tmppoint.String(), string(message)))
	c := ReducedHash(serialize)

	s := new(big.Int).Mul(c, priv.BigInt()) // basically scalar mul add
	s = s.Mod(s, bn256.Order)
	s = s.Add(s, tmpsecret)
	s = s.Mod(s, bn256.Order)

	out := make([]byte, cryptography.SignatureSize)

	sBytes := s.Bytes()
	cBytes := c.Bytes()

	copy(out[(32-len(sBytes)):32], sBytes)
	copy(out[32+(32-len(cBytes)):], cBytes)

	return out, nil
}

func VerifySignature(message, signature, publicKey []byte) bool {

	var u *bn256.G1
	if err := u.DecodeCompressed(publicKey); err != nil {
		return false
	}

	return VerifySignaturePoint(message, signature, u)
}

func VerifySignaturePoint(message, signature []byte, u *bn256.G1) bool {

	if signature == nil {
		return false
	}

	s := new(big.Int).SetBytes(signature[0:32])
	c := new(big.Int).SetBytes(signature[32:64])

	tmppoint := new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, s), new(bn256.G1).ScalarMult(u, new(big.Int).Neg(c)))
	serialize := []byte(fmt.Sprintf("%s%s%s", u.String(), tmppoint.String(), string(message)))

	cCalculated := ReducedHash(serialize)
	if bytes.Equal(c.Bytes(), cCalculated.Bytes()) {
		return true
	}

	return false
}
