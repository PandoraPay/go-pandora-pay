package crypto

import (
	"math/big"
	"pandora-pay/cryptography/bn256"
)

type FieldVector struct {
	vector []*big.Int
}

func NewFieldVector(input []*big.Int) *FieldVector {
	return &FieldVector{vector: input}
}

func NewFieldVectorRandomFilled(capacity int) *FieldVector {
	fv := &FieldVector{vector: make([]*big.Int, capacity, capacity)}
	for i := range fv.vector {
		fv.vector[i] = RandomScalarFixed()
	}
	return fv
}

func (fv *FieldVector) Length() int {
	return len(fv.vector)
}

// slice and return
func (fv *FieldVector) Slice(start, end int) *FieldVector {
	var result FieldVector
	for i := start; i < end; i++ {
		result.vector = append(result.vector, new(big.Int).Set(fv.vector[i]))
	}
	return &result
}

//copy and return
func (fv *FieldVector) Clone() *FieldVector {
	return fv.Slice(0, len(fv.vector))
}

func (fv *FieldVector) Element(index int) *big.Int {
	return fv.vector[index]
}

func (fv *FieldVector) SliceRaw(start, end int) []*big.Int {
	var result FieldVector
	for i := start; i < end; i++ {
		result.vector = append(result.vector, new(big.Int).Set(fv.vector[i]))
	}
	return result.vector
}

func (fv *FieldVector) Flip() *FieldVector {
	var result FieldVector
	for i := range fv.vector {
		result.vector = append(result.vector, new(big.Int).Set(fv.vector[(len(fv.vector)-i)%len(fv.vector)]))
	}
	return &result
}

func (fv *FieldVector) Sum() *big.Int {
	var accumulator big.Int

	for i := range fv.vector {
		var accopy big.Int

		accopy.Add(&accumulator, fv.vector[i])
		accumulator.Mod(&accopy, bn256.Order)
	}

	return &accumulator
}

func (fv *FieldVector) Add(addendum *FieldVector) *FieldVector {
	var result FieldVector

	if len(fv.vector) != len(addendum.vector) {
		panic("mismatched number of elements")
	}

	for i := range fv.vector {
		var ri big.Int
		ri.Mod(new(big.Int).Add(fv.vector[i], addendum.vector[i]), bn256.Order)
		result.vector = append(result.vector, &ri)
	}

	return &result
}

func (gv *FieldVector) AddConstant(c *big.Int) *FieldVector {
	var result FieldVector

	for i := range gv.vector {
		var ri big.Int
		ri.Mod(new(big.Int).Add(gv.vector[i], c), bn256.Order)
		result.vector = append(result.vector, &ri)
	}

	return &result
}

func (fv *FieldVector) Hadamard(exponent *FieldVector) *FieldVector {
	var result FieldVector

	if len(fv.vector) != len(exponent.vector) {
		panic("mismatched number of elements")
	}
	for i := range fv.vector {
		result.vector = append(result.vector, new(big.Int).Mod(new(big.Int).Mul(fv.vector[i], exponent.vector[i]), bn256.Order))
	}

	return &result
}

func (fv *FieldVector) InnerProduct(exponent *FieldVector) *big.Int {
	if len(fv.vector) != len(exponent.vector) {
		panic("mismatched number of elements")
	}

	accumulator := new(big.Int)
	for i := range fv.vector {
		tmp := new(big.Int).Mod(new(big.Int).Mul(fv.vector[i], exponent.vector[i]), bn256.Order)
		accumulator.Add(accumulator, tmp)
		accumulator.Mod(accumulator, bn256.Order)
	}

	return accumulator
}

func (fv *FieldVector) Negate() *FieldVector {
	var result FieldVector
	for i := range fv.vector {
		result.vector = append(result.vector, new(big.Int).Mod(new(big.Int).Neg(fv.vector[i]), bn256.Order))
	}
	return &result
}

func (fv *FieldVector) Times(multiplier *big.Int) *FieldVector {
	var result FieldVector
	for i := range fv.vector {
		res := new(big.Int).Mul(fv.vector[i], multiplier)
		res.Mod(res, bn256.Order)
		result.vector = append(result.vector, res)
	}
	return &result
}

func (fv *FieldVector) Invert() *FieldVector {
	var result FieldVector
	for i := range fv.vector {
		result.vector = append(result.vector, new(big.Int).ModInverse(fv.vector[i], bn256.Order))
	}
	return &result
}

func (fv *FieldVector) Concat(addendum *FieldVector) *FieldVector {
	var result FieldVector
	for i := range fv.vector {
		result.vector = append(result.vector, new(big.Int).Set(fv.vector[i]))
	}

	for i := range addendum.vector {
		result.vector = append(result.vector, new(big.Int).Set(addendum.vector[i]))
	}

	return &result
}

func (fv *FieldVector) Extract(parity bool) *FieldVector {
	var result FieldVector

	remainder := 0
	if parity {
		remainder = 1
	}
	for i := range fv.vector {
		if i%2 == remainder {

			result.vector = append(result.vector, new(big.Int).Set(fv.vector[i]))
		}
	}
	return &result
}
