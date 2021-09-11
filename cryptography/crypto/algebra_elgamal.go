package crypto

import (
	"errors"
	"math/big"
	"pandora-pay/cryptography/bn256"
)

// a ZERO
var ElGamal_ZERO *bn256.G1
var ElGamal_ZERO_string string
var ElGamal_BASE_G *bn256.G1

type ElGamal struct {
	G          *bn256.G1
	Randomness *big.Int
	Left       *bn256.G1
	Right      *bn256.G1
}

func NewElGamal() (p *ElGamal) {
	return &ElGamal{G: global_pedersen_values.G}
}
func CommitElGamal(key *bn256.G1, value *big.Int) *ElGamal {
	e := NewElGamal()
	e.Randomness = RandomScalarFixed()
	e.Left = new(bn256.G1).Add(new(bn256.G1).ScalarMult(e.G, value), new(bn256.G1).ScalarMult(key, e.Randomness))
	e.Right = new(bn256.G1).ScalarMult(G, e.Randomness)
	return e
}

func ConstructElGamal(left, right *bn256.G1) *ElGamal {
	e := NewElGamal()

	if left != nil {
		e.Left = new(bn256.G1).Set(left)
	}
	e.Right = new(bn256.G1).Set(right)
	return e
}

func (e *ElGamal) IsZero() bool {
	if e.Left != nil && e.Right != nil && e.Left.String() == ElGamal_ZERO_string && e.Right.String() == ElGamal_ZERO_string {
		return true
	}
	return false
}

func (e *ElGamal) Add(addendum *ElGamal) *ElGamal {
	if e.Left == nil {
		return ConstructElGamal(nil, new(bn256.G1).Add(e.Right, addendum.Right))
	}
	return ConstructElGamal(new(bn256.G1).Add(e.Left, addendum.Left), new(bn256.G1).Add(e.Right, addendum.Right))
}

func (e *ElGamal) Mul(scalar *big.Int) *ElGamal {
	return ConstructElGamal(new(bn256.G1).ScalarMult(e.Left, scalar), new(bn256.G1).ScalarMult(e.Right, scalar))
}

func (e *ElGamal) Plus(value *big.Int) *ElGamal {
	if e.Right == nil {
		return ConstructElGamal(new(bn256.G1).Add(e.Left, new(bn256.G1).ScalarMult(e.G, value)), nil)
	}
	return ConstructElGamal(new(bn256.G1).Add(e.Left, new(bn256.G1).ScalarMult(e.G, value)), new(bn256.G1).Set(e.Right))
}

func (e *ElGamal) Serialize() (data []byte) {
	if e.Left == nil || e.Right == nil {
		panic("elgamal has nil pointer")
	}
	data = append(data, e.Left.EncodeCompressed()...)
	data = append(data, e.Right.EncodeCompressed()...)
	return data
}

func (e *ElGamal) Deserialize(data []byte) (*ElGamal, error) {

	if len(data) != 66 {
		return nil, errors.New("insufficient buffer size")
	}
	//var left,right *bn256.G1
	left := new(bn256.G1)
	right := new(bn256.G1)

	if err := left.DecodeCompressed(data[:33]); err != nil {
		return nil, err
	}
	if err := right.DecodeCompressed(data[33:66]); err != nil {
		return nil, err
	}
	e = ConstructElGamal(left, right)
	return e, nil
}

func (e *ElGamal) Neg() *ElGamal {

	var left, right *bn256.G1
	if e.Left != nil {
		left = new(bn256.G1).Neg(e.Left)
	}
	if e.Right != nil {
		right = new(bn256.G1).Neg(e.Right)
	}
	return ConstructElGamal(left, right)
}

type ElGamalVector struct {
	vector []*ElGamal
}

func (e *ElGamalVector) MultiExponentiate(exponents *FieldVector) *ElGamal {
	accumulator := ConstructElGamal(ElGamal_ZERO, ElGamal_ZERO)
	for i := range exponents.vector {
		accumulator = accumulator.Add(e.vector[i].Mul(exponents.vector[i]))
	}
	return accumulator
}

func (e *ElGamalVector) Sum() *ElGamal {
	r := ConstructElGamal(ElGamal_ZERO, ElGamal_ZERO)
	for i := range e.vector {
		r = r.Add(e.vector[i])
	}
	return r
}

func (e *ElGamalVector) Add(other *ElGamalVector) *ElGamalVector {
	var r ElGamalVector
	for i := range e.vector {
		r.vector = append(r.vector, ConstructElGamal(e.vector[i].Left, e.vector[i].Right))
	}

	for i := range other.vector {
		r.vector[i] = r.vector[i].Add(other.vector[i])
	}
	return &r
}

func (e *ElGamalVector) Hadamard(exponents FieldVector) *ElGamalVector {
	var r ElGamalVector
	for i := range e.vector {
		r.vector = append(r.vector, ConstructElGamal(e.vector[i].Left, e.vector[i].Right))
	}

	for i := range exponents.vector {
		r.vector[i] = r.vector[i].Mul(exponents.vector[i])
	}
	return &r
}

func (e *ElGamalVector) Times(scalar *big.Int) *ElGamalVector {
	var r ElGamalVector
	for i := range e.vector {
		r.vector = append(r.vector, e.vector[i].Mul(scalar))
	}
	return &r
}
