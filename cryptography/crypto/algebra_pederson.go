package crypto

import (
	"fmt"
	"math/big"
	"pandora-pay/cryptography/bn256"
)

var G *bn256.G1
var global_pedersen_values PedersenVectorCommitment

func init() {
	var zeroes [64]byte
	var gs, hs []*bn256.G1

	global_pedersen_values.G = HashToPoint(HashtoNumber([]byte(PROTOCOL_CONSTANT + "G"))) // this is same as mybase or vice-versa
	global_pedersen_values.H = HashToPoint(HashtoNumber([]byte(PROTOCOL_CONSTANT + "H")))

	global_pedersen_values.GSUM = new(bn256.G1)
	global_pedersen_values.GSUM.Unmarshal(zeroes[:])

	for i := 0; i < 128; i++ {
		gs = append(gs, HashToPoint(HashtoNumber(append([]byte(PROTOCOL_CONSTANT+"G"), hextobytes(makestring64(fmt.Sprintf("%x", i)))...))))
		hs = append(hs, HashToPoint(HashtoNumber(append([]byte(PROTOCOL_CONSTANT+"H"), hextobytes(makestring64(fmt.Sprintf("%x", i)))...))))

		global_pedersen_values.GSUM = new(bn256.G1).Add(global_pedersen_values.GSUM, gs[i])
	}
	global_pedersen_values.Gs = NewPointVector(gs)
	global_pedersen_values.Hs = NewPointVector(hs)

	// also initialize elgamal_zero
	ElGamal_ZERO = new(bn256.G1).ScalarMult(global_pedersen_values.G, new(big.Int).SetUint64(0))
	ElGamal_ZERO_string = ElGamal_ZERO.String()
	ElGamal_BASE_G = global_pedersen_values.G
	G = global_pedersen_values.G
	((*bn256.G1)(&GPoint)).Set(G) // setup base point

	//   fmt.Printf("basepoint %s on %x\n", G.String(), G.Marshal())
}

type PedersenCommitmentNew struct {
	G          *bn256.G1
	H          *bn256.G1
	Randomness *big.Int
	Result     *bn256.G1
}

func NewPedersenCommitmentNew() (p *PedersenCommitmentNew) {
	return &PedersenCommitmentNew{G: global_pedersen_values.G, H: global_pedersen_values.H}
}

// commit a specific value to specific bases
func (p *PedersenCommitmentNew) Commit(value *big.Int) *PedersenCommitmentNew {
	p.Randomness = RandomScalarFixed()
	point := new(bn256.G1).Add(new(bn256.G1).ScalarMult(p.G, value), new(bn256.G1).ScalarMult(p.H, p.Randomness))
	p.Result = new(bn256.G1).Set(point)
	return p
}

type PedersenVectorCommitment struct {
	G    *bn256.G1
	H    *bn256.G1
	GSUM *bn256.G1

	Gs         *PointVector
	Hs         *PointVector
	Randomness *big.Int
	Result     *bn256.G1

	gvalues *FieldVector
	hvalues *FieldVector
}

func NewPedersenVectorCommitment() (p *PedersenVectorCommitment) {
	p = &PedersenVectorCommitment{}
	*p = global_pedersen_values
	return
}

// commit a specific value to specific bases
func (p *PedersenVectorCommitment) Commit(gvalues, hvalues *FieldVector) *PedersenVectorCommitment {

	p.Randomness = RandomScalarFixed()
	point := new(bn256.G1).ScalarMult(p.H, p.Randomness)
	point = new(bn256.G1).Add(point, p.Gs.MultiExponentiate(gvalues))
	point = new(bn256.G1).Add(point, p.Hs.MultiExponentiate(hvalues))

	p.Result = new(bn256.G1).Set(point)
	return p
}
