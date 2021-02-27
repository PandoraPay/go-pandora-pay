package block

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"pandora-pay/blockchain/account"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/config"
	"pandora-pay/config/reward"
	"pandora-pay/crypto"
	"pandora-pay/crypto/ecdsa"
	"pandora-pay/helpers"
)

type Block struct {
	BlockHeader
	MerkleHash helpers.Hash

	PrevHash       helpers.Hash
	PrevKernelHash helpers.Hash

	Timestamp uint64

	StakingAmount uint64

	Forger    [33]byte // 33 byte public key
	Signature [65]byte // 65 byte signature
}

func (blk *Block) IncludeBlock(acs *accounts.Accounts) (err error) {

	forgerPublicKeyHash := blk.GetForgerPublicKeyHash()
	var acc *account.Account

	if acc, err = acs.GetAccountEvenEmpty(string(forgerPublicKeyHash[:])); err != nil {
		return
	}

	reward := config.ConvertToUnits(reward.GetRewardAt(blk.Height))
	if err = acc.AddBalance(true, reward, config.NATIVE_CURRENCY); err != nil {
		return
	}

	if err = acs.UpdateAccount(string(forgerPublicKeyHash[:]), acc); err != nil {
		return
	}

	return
}

func (blk *Block) RemoveBlock(acs *accounts.Accounts) (err error) {

	forgerPublicKeyHash := blk.GetForgerPublicKeyHash()
	var acc *account.Account

	if acc, err = acs.GetAccount(string(forgerPublicKeyHash[:])); err != nil {
		return
	}

	if err = acc.AddBalance(false, reward.GetRewardAt(blk.Height), config.NATIVE_CURRENCY); err != nil {
		return
	}

	if err = acs.UpdateAccount(string(forgerPublicKeyHash[:]), acc); err != nil {
		return
	}

	return
}

func (blk *Block) GetForgerPublicKeyHash() [20]byte {
	return *helpers.Byte20(crypto.ComputePublicKeyHash(blk.Forger[:]))
}

func (blk *Block) ComputeHash() helpers.Hash {
	return crypto.SHA3Hash(blk.Serialize())
}

func (blk *Block) ComputeKernelHashOnly() helpers.Hash {
	return crypto.SHA3Hash(blk.SerializeBlock(true, false))
}

func (blk *Block) ComputeKernelHash() helpers.Hash {

	hash := blk.ComputeKernelHashOnly()

	if blk.Height == 0 {
		return hash
	}

	number := new(big.Int).Div(new(big.Int).SetBytes(hash[:]), new(big.Int).SetUint64(blk.StakingAmount))

	buf := number.Bytes()
	var finalHash helpers.Hash
	copy(finalHash[helpers.HashSize-len(buf):], buf)

	return finalHash
}

func (blk *Block) SerializeForSigning() helpers.Hash {
	return crypto.SHA3Hash(blk.SerializeBlock(false, false))
}

func (blk *Block) VerifySignature() bool {
	hash := blk.SerializeForSigning()
	return ecdsa.VerifySignature(blk.Forger[:], hash[:], blk.Signature[0:64])
}

func (blk *Block) SerializeBlock(kernelHash bool, inclSignature bool) []byte {

	var serialized bytes.Buffer
	temp := make([]byte, binary.MaxVarintLen64)

	blk.BlockHeader.Serialize(&serialized, temp)

	if !kernelHash {
		serialized.Write(blk.MerkleHash[:])
		serialized.Write(blk.PrevHash[:])
	}

	serialized.Write(blk.PrevKernelHash[:])

	if !kernelHash {
		n := binary.PutUvarint(temp, blk.Timestamp)
		serialized.Write(temp[:n])

		n = binary.PutUvarint(temp, blk.StakingAmount)
		serialized.Write(temp[:n])
	}

	serialized.Write(blk.Forger[:])

	if inclSignature {
		serialized.Write(blk.Signature[:])
	}

	return serialized.Bytes()
}

func (blk *Block) Serialize() []byte {
	return blk.SerializeBlock(false, true)
}

func (blk *Block) Deserialize(buf []byte) (out []byte, err error) {

	if buf, err = blk.BlockHeader.Deserialize(buf); err != nil {
		return
	}

	if blk.MerkleHash, buf, err = helpers.DeserializeHash(buf, helpers.HashSize); err != nil {
		return
	}

	if blk.PrevHash, buf, err = helpers.DeserializeHash(buf, helpers.HashSize); err != nil {
		return
	}

	if blk.PrevKernelHash, buf, err = helpers.DeserializeHash(buf, helpers.HashSize); err != nil {
		return
	}

	if blk.Timestamp, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	if blk.StakingAmount, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	var data []byte
	if data, buf, err = helpers.DeserializeBuffer(buf, 33); err != nil {
		return
	}
	blk.Forger = *helpers.Byte33(data)

	if data, buf, err = helpers.DeserializeBuffer(buf, 65); err != nil {
		return
	}
	blk.Signature = *helpers.Byte65(data)

	out = buf
	return
}
