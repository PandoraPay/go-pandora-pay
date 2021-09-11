package crypto

import (
	"fmt"
	"math/big"
)

type BNRed big.Int

func RandomScalarBNRed() *BNRed {
	return (*BNRed)(RandomScalar())
}

// converts  big.Int to BNRed
func GetBNRed(x *big.Int) *BNRed {
	result := new(BNRed)

	((*big.Int)(result)).Set(x)
	return result
}

// convert BNRed to BigInt
func (x *BNRed) BigInt() *big.Int {
	return new(big.Int).Set(((*big.Int)(x)))
}

func (x *BNRed) SetBytes(buf []byte) *BNRed {
	((*big.Int)(x)).SetBytes(buf)
	return x
}

func (x *BNRed) ToBytes() []byte {
	return ((*big.Int)(x)).Bytes()
}

func (x *BNRed) String() string {
	return ((*big.Int)(x)).Text(16)
}

func (x *BNRed) Text(base int) string {
	return ((*big.Int)(x)).Text(base)
}

func (x *BNRed) MarshalText() ([]byte, error) {
	return []byte(((*big.Int)(x)).Text(16)), nil
}

func (x *BNRed) UnmarshalText(text []byte) error {
	_, err := fmt.Sscan("0x"+string(text), ((*big.Int)(x)))
	return err
}
