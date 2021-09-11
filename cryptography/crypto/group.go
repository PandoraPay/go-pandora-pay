package crypto

import (
	"encoding/hex"
	"math/big"
	"pandora-pay/cryptography/bn256"
)

type Point bn256.G1

var GPoint Point

// ScalarMult with chainable API
func (p *Point) ScalarMult(r *BNRed) (result *Point) {
	result = new(Point)
	((*bn256.G1)(result)).ScalarMult(((*bn256.G1)(p)), ((*big.Int)(r)))
	return result
}

func (p *Point) EncodeCompressed() []byte {
	return ((*bn256.G1)(p)).EncodeCompressed()
}

func (p *Point) DecodeCompressed(i []byte) error {
	return ((*bn256.G1)(p)).DecodeCompressed(i)
}

func (p *Point) G1() *bn256.G1 {
	return ((*bn256.G1)(p))
}
func (p *Point) Set(x *Point) *Point {
	return ((*Point)(((*bn256.G1)(p)).Set(((*bn256.G1)(x)))))
}

func (p *Point) String() string {
	return string(((*bn256.G1)(p)).EncodeCompressed())
}

func (p *Point) StringHex() string {
	return string(hex.EncodeToString(((*bn256.G1)(p)).EncodeCompressed()))
}

func (p *Point) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(((*bn256.G1)(p)).EncodeCompressed())), nil
}

func (p *Point) UnmarshalText(text []byte) error {

	tmp, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	return p.DecodeCompressed(tmp)
}
