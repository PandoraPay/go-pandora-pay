package crypto

import (
	"math/big"
	"pandora-pay/cryptography/bn256"
)

// ToDO evaluate others curves such as BLS12 used by zcash, also BLS24 or others , providing ~200 bits of security , may be required for long time use( say 50 years)
type PointVector struct {
	vector []*bn256.G1
}

func NewPointVector(input []*bn256.G1) *PointVector {
	return &PointVector{vector: input}
}

func (gv *PointVector) Length() int {
	return len(gv.vector)
}

// slice and return
func (gv *PointVector) Slice(start, end int) *PointVector {
	var result PointVector
	for i := start; i < end; i++ {
		var ri bn256.G1
		ri.Set(gv.vector[i])
		result.vector = append(result.vector, &ri)
	}
	return &result
}

func (gv *PointVector) Commit(exponent []*big.Int) *bn256.G1 {
	var accumulator, zero bn256.G1
	var zeroes [64]byte
	accumulator.Unmarshal(zeroes[:]) // obtain zero element, this should be static and
	zero.Unmarshal(zeroes[:])

	accumulator.ScalarMult(G, new(big.Int))

	//fmt.Printf("zero %s\n", accumulator.String())

	if len(gv.vector) != len(exponent) {
		panic("mismatched number of elements")
	}
	for i := range gv.vector { // TODO a bug exists somewhere deep here
		var tmp, accopy bn256.G1
		tmp.ScalarMult(gv.vector[i], exponent[i])

		accopy.Set(&accumulator)
		accumulator.Add(&accopy, &tmp)
	}

	return &accumulator
}

func (gv *PointVector) Sum() *bn256.G1 {
	var accumulator bn256.G1

	accumulator.ScalarMult(G, new(big.Int)) // set it to zero

	for i := range gv.vector {
		var accopy bn256.G1

		accopy.Set(&accumulator)
		accumulator.Add(&accopy, gv.vector[i])
	}

	return &accumulator
}

func (gv *PointVector) Add(addendum *PointVector) *PointVector {
	var result PointVector

	if len(gv.vector) != len(addendum.vector) {
		panic("mismatched number of elements")
	}

	for i := range gv.vector {
		var ri bn256.G1

		ri.Add(gv.vector[i], addendum.vector[i])
		result.vector = append(result.vector, &ri)
	}

	return &result
}

func (gv *PointVector) Hadamard(exponent []*big.Int) *PointVector {
	var result PointVector

	if len(gv.vector) != len(exponent) {
		panic("mismatched number of elements")
	}
	for i := range gv.vector {
		var ri bn256.G1
		ri.ScalarMult(gv.vector[i], exponent[i])
		result.vector = append(result.vector, &ri)

	}

	return &result
}

func (gv *PointVector) Negate() *PointVector {
	var result PointVector
	for i := range gv.vector {
		var ri bn256.G1
		ri.Neg(gv.vector[i])
		result.vector = append(result.vector, &ri)
	}
	return &result
}

func (gv *PointVector) Times(multiplier *big.Int) *PointVector {
	var result PointVector
	for i := range gv.vector {
		var ri bn256.G1
		ri.ScalarMult(gv.vector[i], multiplier)
		result.vector = append(result.vector, &ri)
	}
	return &result
}

func (gv *PointVector) Extract(parity bool) *PointVector {
	var result PointVector

	remainder := 0
	if parity {
		remainder = 1
	}
	for i := range gv.vector {
		if i%2 == remainder {
			var ri bn256.G1
			ri.Set(gv.vector[i])
			result.vector = append(result.vector, &ri)
		}
	}
	return &result
}

func (gv *PointVector) Concat(addendum *PointVector) *PointVector {
	var result PointVector
	for i := range gv.vector {
		var ri bn256.G1
		ri.Set(gv.vector[i])
		result.vector = append(result.vector, &ri)
	}

	for i := range addendum.vector {
		var ri bn256.G1
		ri.Set(addendum.vector[i])
		result.vector = append(result.vector, &ri)
	}

	return &result
}

func (pv *PointVector) MultiExponentiate(fv *FieldVector) *bn256.G1 {
	var accumulator bn256.G1

	accumulator.ScalarMult(G, new(big.Int)) // set it to zero

	for i := range fv.vector {
		var accopy bn256.G1

		accopy.Set(&accumulator)
		accumulator.Add(&accopy, new(bn256.G1).ScalarMult(pv.vector[i], fv.vector[i]))
	}

	return &accumulator
}
