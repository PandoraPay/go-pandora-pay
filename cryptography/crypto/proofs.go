package crypto

import (
	"errors"
	"math/big"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations/transaction_zether_registration"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/helpers"
)

type Statement struct {
	RingSize      int
	CLn           []*bn256.G1 //bloomed
	CRn           []*bn256.G1 //bloomed
	Publickeylist []*bn256.G1 //bloomed
	C             []*bn256.G1 // commitments
	D             *bn256.G1
	Fee           uint64
}

type Witness struct {
	SecretKey      *big.Int
	R              *big.Int
	TransferAmount uint64 // total value being transferred
	Balance        uint64 // whatever is the the amount left after transfer
	Index          []int  // index of sender in the public key list
}

func (s *Statement) SerializeRingSize(w *helpers.BufferWriter) {
	pow, err := GetPowerof2(len(s.C))
	if err != nil {
		panic(err)
	}

	w.WriteByte(byte(pow)) // len(s.Publickeylist) is always power of 2
}

func (s *Statement) Serialize(w *helpers.BufferWriter, payloadRegistrations []*transaction_zether_registration.TransactionZetherDataRegistration) {

	w.WriteUvarint(s.Fee)
	w.Write(s.D.EncodeCompressed())

	for i := 0; i < len(s.C); i++ {
		w.Write(s.Publickeylist[i].EncodeCompressed()) //can be bloomed
		w.Write(s.C[i].EncodeCompressed())
		if payloadRegistrations[i].RegistrationType == transaction_zether_registration.REGISTERED_ACCOUNT {
			w.Write(s.CLn[i].EncodeCompressed()) //can be bloomed
			w.Write(s.CRn[i].EncodeCompressed()) //can be bloomed
		}
	}

}

func (s *Statement) DeserializeRingSize(r *helpers.BufferReader) (byte, int, error) {

	length, err := r.ReadByte()
	if err != nil {
		return 0, 0, nil
	}

	if length > 8 || length < 1 {
		return 0, 0, errors.New("Invalid Ring Length Power")
	}

	s.RingSize = 1 << length

	return length, s.RingSize, nil
}

func (s *Statement) Deserialize(r *helpers.BufferReader, payloadRegistrations []*transaction_zether_registration.TransactionZetherDataRegistration) (err error) {

	if s.Fee, err = r.ReadUvarint(); err != nil {
		return
	}

	if s.D, err = r.ReadBN256G1(); err != nil {
		return
	}

	s.CLn = make([]*bn256.G1, s.RingSize)
	s.CRn = make([]*bn256.G1, s.RingSize)
	s.Publickeylist = make([]*bn256.G1, s.RingSize)
	s.C = make([]*bn256.G1, s.RingSize)
	for i := 0; i < s.RingSize; i++ {
		if s.Publickeylist[i], err = r.ReadBN256G1(); err != nil {
			return
		}
		if s.C[i], err = r.ReadBN256G1(); err != nil {
			return
		}
		if payloadRegistrations[i].RegistrationType == transaction_zether_registration.REGISTERED_ACCOUNT {
			if s.CLn[i], err = r.ReadBN256G1(); err != nil {
				return
			}
			if s.CRn[i], err = r.ReadBN256G1(); err != nil {
				return
			}
		} else {
			var acckey Point
			if err = acckey.DecodeCompressed(s.Publickeylist[i].EncodeCompressed()); err != nil {
				return
			}
			balance := ConstructElGamal(acckey.G1(), ElGamal_BASE_G)

			var left, right bn256.G1
			left.Add(balance.Left, s.C[i])
			s.CLn[i] = &left

			right.Add(balance.Right, s.D)
			s.CRn[i] = &right
		}
	}

	return nil
}
