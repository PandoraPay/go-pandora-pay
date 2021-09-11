package crypto

import (
	"math/big"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/helpers"
)

type Statement struct {
	RingSize      uint64
	CLn           []*bn256.G1
	CRn           []*bn256.G1
	Publickeylist []*bn256.G1
	C             []*bn256.G1 // commitments
	D             *bn256.G1
	Fees          uint64

	Roothash []byte // note roothash contains the merkle root hash of chain, when it was build
}

type Witness struct {
	SecretKey      *big.Int
	R              *big.Int
	TransferAmount uint64 // total value being transferred
	Balance        uint64 // whatever is the the amount left after transfer
	Index          []int  // index of sender in the public key list
}

func (s *Statement) Serialize(w *helpers.BufferWriter) {

	pow, err := GetPowerof2(len(s.Publickeylist))
	if err != nil {
		panic(err)
	}
	w.WriteByte(byte(pow)) // len(s.Publickeylist) is always power of 2
	w.WriteUvarint(s.Fees)
	w.Write(s.D.EncodeCompressed())

	for i := 0; i < len(s.Publickeylist); i++ {
		//     w.Write( s.CLn[i].EncodeCompressed()) /// this is expanded from graviton store
		//     w.Write( s.CRn[i].EncodeCompressed())  /// this is expanded from graviton store
		w.Write(s.Publickeylist[i].EncodeCompressed()) /// this is expanded from graviton store
		w.Write(s.C[i].EncodeCompressed())
	}

	w.Write(s.Roothash)
}

func (s *Statement) Deserialize(r *helpers.BufferReader) (err error) {

	var p bn256.G1
	var bufp []byte

	length, err := r.ReadByte()
	if err != nil {
		return
	}
	s.RingSize = 1 << length

	if s.Fees, err = r.ReadUvarint(); err != nil {
		return
	}

	if bufp, err = r.ReadBytes(33); err != nil {
		return
	}
	p = bn256.G1{}
	if err = p.DecodeCompressed(bufp[:]); err != nil {
		return err
	}
	s.D = &p

	for i := 0; i < int(s.RingSize); i++ {

		if bufp, err = r.ReadBytes(33); err != nil {
			return
		}
		p = bn256.G1{}
		if err = p.DecodeCompressed(bufp[:]); err != nil {
			return err
		}
		s.Publickeylist = append(s.Publickeylist, &p)

		if bufp, err = r.ReadBytes(33); err != nil {
			return
		}

		p = bn256.G1{}
		if err = p.DecodeCompressed(bufp[:]); err != nil {
			return err
		}
		s.C = append(s.C, &p)
	}

	if s.Roothash, err = r.ReadBytes(32); err != nil {
		return
	}

	return nil

}
