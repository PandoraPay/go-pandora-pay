package crypto

import (
	"math/big"
	"pandora-pay/cryptography/bn256"
)

type Polynomial struct {
	coefficients []*big.Int
}

func NewPolynomial(input []*big.Int) *Polynomial {
	if input == nil {
		return &Polynomial{coefficients: []*big.Int{new(big.Int).SetInt64(1)}}
	}
	return &Polynomial{coefficients: input}
}

func (p *Polynomial) Length() int {
	return len(p.coefficients)
}

func (p *Polynomial) Mul(m *Polynomial) *Polynomial {
	var product []*big.Int
	for i := range p.coefficients {
		product = append(product, new(big.Int).Mod(new(big.Int).Mul(p.coefficients[i], m.coefficients[0]), bn256.Order))
	}
	product = append(product, new(big.Int)) // add 0 element

	if m.coefficients[1].IsInt64() && m.coefficients[1].Int64() == 1 {
		for i := range product {
			if i > 0 {
				tmp := new(big.Int).Add(product[i], p.coefficients[i-1])

				product[i] = new(big.Int).Mod(tmp, bn256.Order)

			} else { // do nothing

			}
		}
	}
	return NewPolynomial(product)
}

type dummy struct {
	list [][]*big.Int
}

func RecursivePolynomials(list [][]*big.Int, accum *Polynomial, a, b []*big.Int) (rlist [][]*big.Int) {
	var d dummy
	d.recursivePolynomialsinternal(accum, a, b)

	return d.list
}

func (d *dummy) recursivePolynomialsinternal(accum *Polynomial, a, b []*big.Int) {
	if len(a) == 0 {
		d.list = append(d.list, accum.coefficients)
		return
	}

	atop := a[len(a)-1]
	btop := b[len(b)-1]

	left := NewPolynomial([]*big.Int{new(big.Int).Mod(new(big.Int).Neg(atop), bn256.Order), new(big.Int).Mod(new(big.Int).Sub(new(big.Int).SetInt64(1), btop), bn256.Order)})
	right := NewPolynomial([]*big.Int{atop, btop})

	d.recursivePolynomialsinternal(accum.Mul(left), a[:len(a)-1], b[:len(b)-1])
	d.recursivePolynomialsinternal(accum.Mul(right), a[:len(a)-1], b[:len(b)-1])
}
