package crypto

import (
	"math/big"
	"pandora-pay/cryptography/bn256"
)

func (proof *Proof) GetA_t(txid []byte) []byte {

	statementhash := reducedhash(txid[:])
	var input []byte
	input = append(input, ConvertBigIntToByte(statementhash)...)
	input = append(input, proof.BA.Marshal()...)
	input = append(input, proof.BS.Marshal()...)
	input = append(input, proof.A.Marshal()...)
	input = append(input, proof.B.Marshal()...)

	anonsupportv := reducedhash(input)

	anonsupportw := proof.hashmash1(anonsupportv)

	protsupporty := reducedhash(ConvertBigIntToByte(anonsupportw))

	var protsupportys []*big.Int
	protsupportys = append(protsupportys, new(big.Int).SetUint64(1))
	protsupportk := new(big.Int).SetUint64(1)
	for i := 1; i < 128; i++ {
		protsupportys = append(protsupportys, new(big.Int).Mod(new(big.Int).Mul(protsupportys[i-1], protsupporty), bn256.Order))
		protsupportk = new(big.Int).Mod(new(big.Int).Add(protsupportk, protsupportys[i]), bn256.Order)
	}

	protsupportz := reducedhash(ConvertBigIntToByte(protsupporty))
	protsupportzs := []*big.Int{new(big.Int).Exp(protsupportz, new(big.Int).SetUint64(2), bn256.Order), new(big.Int).Exp(protsupportz, new(big.Int).SetUint64(3), bn256.Order)}

	protsupportzSum := new(big.Int).Mod(new(big.Int).Add(protsupportzs[0], protsupportzs[1]), bn256.Order)
	protsupportzSum = new(big.Int).Mod(new(big.Int).Mul(new(big.Int).Set(protsupportzSum), protsupportz), bn256.Order)

	z_z0 := new(big.Int).Mod(new(big.Int).Sub(protsupportz, protsupportzs[0]), bn256.Order)
	protsupportk = new(big.Int).Mod(new(big.Int).Mul(protsupportk, z_z0), bn256.Order)

	proof_2_64, _ := new(big.Int).SetString("18446744073709551616", 10)
	zsum_pow := new(big.Int).Mod(new(big.Int).Mul(protsupportzSum, proof_2_64), bn256.Order)
	zsum_pow = new(big.Int).Mod(new(big.Int).Sub(zsum_pow, protsupportzSum), bn256.Order)
	protsupportk = new(big.Int).Mod(new(big.Int).Sub(protsupportk, zsum_pow), bn256.Order)

	protsupportt := new(big.Int).Mod(new(big.Int).Sub(proof.that, protsupportk), bn256.Order) // t = tHat - delta(y, z)

	x := new(big.Int)

	{
		var input []byte
		input = append(input, ConvertBigIntToByte(protsupportz)...) // tie intermediates/commit
		input = append(input, proof.T_1.Marshal()...)
		input = append(input, proof.T_2.Marshal()...)
		x = reducedhash(input)
	}

	xsq := new(big.Int).Mod(new(big.Int).Mul(x, x), bn256.Order)
	protsupporttEval := new(bn256.G1).ScalarMult(proof.T_1, x)
	protsupporttEval.Add(new(bn256.G1).Set(protsupporttEval), new(bn256.G1).ScalarMult(proof.T_2, xsq))

	proof_s_b_neg := new(big.Int).Mod(new(big.Int).Neg(proof.s_b), bn256.Order)
	wPow := new(big.Int).SetUint64(1)

	m := proof.f.Length() / 2
	for i := 0; i < m; i++ {
		wPow = new(big.Int).Mod(new(big.Int).Mul(wPow, anonsupportw), bn256.Order)
	}

	A_t := new(bn256.G1).ScalarMult(gparams.G, protsupportt)
	A_t.Add(new(bn256.G1).Set(A_t), new(bn256.G1).Neg(protsupporttEval))
	A_t = new(bn256.G1).ScalarMult(A_t, new(big.Int).Mod(new(big.Int).Mul(proof.c, wPow), bn256.Order))
	A_t.Add(new(bn256.G1).Set(A_t), new(bn256.G1).ScalarMult(gparams.H, proof.s_tau))
	A_t.Add(new(bn256.G1).Set(A_t), new(bn256.G1).ScalarMult(gparams.G, proof_s_b_neg))

	return A_t.Marshal()
}
