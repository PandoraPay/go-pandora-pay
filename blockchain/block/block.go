package block

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/accounts/account/dpos"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
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

	DelegatedPublicKey [33]byte //33 byte public key. It IS NOT included in the kernel hash
	Forger             [20]byte // 20 byte public key hash
	Signature          [65]byte // 65 byte signature
}

func (blk *Block) IncludeBlock(acs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	var acc *account.Account

	if acc, err = acs.GetAccountEvenEmpty(blk.Forger); err != nil {
		return
	}

	//for genesis block
	if blk.Height == 0 && !acc.HasDelegatedStake() {
		acc.DelegatedStakeVersion = 1
		acc.DelegatedStake = new(dpos.DelegatedStake)
		acc.DelegatedStake.DelegatedPublicKey = blk.DelegatedPublicKey
	}

	reward := reward.GetRewardAt(blk.Height)
	acc.AddReward(true, reward, blk.Height)
	acs.UpdateAccount(blk.Forger, acc)

	var tok *token.Token
	if tok, err = toks.GetToken(config.NATIVE_TOKEN_FULL); err != nil {
		return
	}
	if err = tok.AddSupply(true, reward); err != nil {
		return
	}
	toks.UpdateToken(config.NATIVE_TOKEN_FULL, tok)

	return
}

func (blk *Block) RemoveBlock(acs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	var acc *account.Account

	if acc, err = acs.GetAccount(blk.Forger); err != nil {
		return
	}

	reward := reward.GetRewardAt(blk.Height)
	acc.AddReward(false, reward, blk.Height)

	acs.UpdateAccount(blk.Forger, acc)

	var tok *token.Token
	if tok, err = toks.GetToken(config.NATIVE_TOKEN_FULL); err != nil {
		return
	}
	if err = tok.AddSupply(false, reward); err != nil {
		return
	}
	toks.UpdateToken(config.NATIVE_TOKEN_FULL, tok)

	return
}

func (blk *Block) GetDelegatePublicKeyHash() [20]byte {
	return *helpers.Byte20(crypto.ComputePublicKeyHash(blk.Forger[:]))
}

func (blk *Block) ComputeHash() helpers.Hash {
	return crypto.SHA3Hash(blk.Serialize())
}

func (blk *Block) ComputeKernelHashOnly() helpers.Hash {
	out := blk.SerializeBlock(true, false)
	return crypto.SHA3Hash(out)
}

func (blk *Block) ComputeKernelHash() helpers.Hash {

	hash := blk.ComputeKernelHashOnly()

	if blk.Height == 0 {
		return hash
	}

	return crypto.ComputeKernelHash(hash, blk.StakingAmount)
}

func (blk *Block) SerializeForSigning() helpers.Hash {
	return crypto.SHA3Hash(blk.SerializeBlock(false, false))
}

func (blk *Block) VerifySignature() bool {
	hash := blk.SerializeForSigning()
	return ecdsa.VerifySignature(blk.DelegatedPublicKey[:], hash[:], blk.Signature[0:64])
}

func (blk *Block) SerializeBlock(kernelHash bool, inclSignature bool) []byte {

	writer := helpers.NewBufferWriter()

	blk.BlockHeader.Serialize(writer)

	if !kernelHash {
		writer.Write(blk.MerkleHash[:])
		writer.Write(blk.PrevHash[:])
	}

	writer.Write(blk.PrevKernelHash[:])

	if !kernelHash {

		writer.WriteUint64(blk.StakingAmount)
		writer.Write(blk.DelegatedPublicKey[:])
	}

	writer.WriteUint64(blk.Timestamp)

	writer.Write(blk.Forger[:])

	if inclSignature {
		writer.Write(blk.Signature[:])
	}

	return writer.Bytes()
}

func (blk *Block) Serialize() []byte {
	return blk.SerializeBlock(false, true)
}

func (blk *Block) Deserialize(reader *helpers.BufferReader) (err error) {

	if err = blk.BlockHeader.Deserialize(reader); err != nil {
		return
	}

	if blk.MerkleHash, err = reader.ReadHash(); err != nil {
		return
	}

	if blk.PrevHash, err = reader.ReadHash(); err != nil {
		return
	}

	if blk.PrevKernelHash, err = reader.ReadHash(); err != nil {
		return
	}

	if blk.StakingAmount, err = reader.ReadUvarint(); err != nil {
		return
	}

	var data []byte
	if data, err = reader.ReadBytes(33); err != nil {
		return
	}
	blk.DelegatedPublicKey = *helpers.Byte33(data)

	if blk.Timestamp, err = reader.ReadUvarint(); err != nil {
		return
	}

	if data, err = reader.ReadBytes(20); err != nil {
		return
	}
	blk.Forger = *helpers.Byte20(data)

	if data, err = reader.ReadBytes(65); err != nil {
		return
	}
	blk.Signature = *helpers.Byte65(data)

	return
}
