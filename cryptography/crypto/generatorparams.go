package crypto

import (
	"fmt"
	"math/big"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
)

func NewGeneratorParams(count int) *GeneratorParams {
	GP := &GeneratorParams{}
	var zeroes [64]byte

	GP.G = HashToPoint(HashtoNumber([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT + "G"))) // this is same as mybase or vice-versa
	GP.H = HashToPoint(HashtoNumber([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT + "H")))

	var gs, hs []*bn256.G1

	GP.GSUM = new(bn256.G1)
	GP.GSUM.Unmarshal(zeroes[:])

	for i := 0; i < count; i++ {
		gs = append(gs, HashToPoint(HashtoNumber(append([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT+"G"), hextobytes(makestring64(fmt.Sprintf("%x", i)))...))))
		hs = append(hs, HashToPoint(HashtoNumber(append([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT+"H"), hextobytes(makestring64(fmt.Sprintf("%x", i)))...))))

		GP.GSUM = new(bn256.G1).Add(GP.GSUM, gs[i])
	}
	GP.Gs = NewPointVector(gs)
	GP.Hs = NewPointVector(hs)

	return GP
}

func NewGeneratorParams3(h *bn256.G1, gs, hs *PointVector) *GeneratorParams {
	GP := &GeneratorParams{}

	GP.G = HashToPoint(HashtoNumber([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT + "G"))) // this is same as mybase or vice-versa
	GP.H = h
	GP.Gs = gs
	GP.Hs = hs
	return GP
}

func (gp *GeneratorParams) Commit(blind *big.Int, gexps, hexps *FieldVector) *bn256.G1 {
	result := new(bn256.G1).ScalarMult(gp.H, blind)
	for i := range gexps.vector {
		result = new(bn256.G1).Add(result, new(bn256.G1).ScalarMult(gp.Gs.vector[i], gexps.vector[i]))
	}
	if hexps != nil {
		for i := range hexps.vector {
			result = new(bn256.G1).Add(result, new(bn256.G1).ScalarMult(gp.Hs.vector[i], hexps.vector[i]))
		}
	}
	return result
}
