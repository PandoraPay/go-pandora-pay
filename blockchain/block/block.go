package block

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account/dpos"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/config"
	"pandora-pay/config/reward"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type Block struct {
	BlockHeader
	MerkleHash cryptography.Hash

	PrevHash       cryptography.Hash
	PrevKernelHash cryptography.Hash

	Timestamp uint64

	StakingAmount uint64

	DelegatedPublicKey [33]byte //33 byte public key. It IS NOT included in the kernel hash
	Forger             [20]byte // 20 byte public key hash
	Signature          [65]byte // 65 byte signature
}

func (blk *Block) Validate() {
	blk.BlockHeader.Validate()
}

func (blk *Block) Verify() {
	if blk.VerifySignature() != true {
		panic("Forger Signature is invalid!")
	}
}

func (blk *Block) IncludeBlock(acs *accounts.Accounts, toks *tokens.Tokens, allFees map[string]uint64) {

	reward := reward.GetRewardAt(blk.Height)
	acc := acs.GetAccountEvenEmpty(blk.Forger)
	acc.RefreshDelegatedStake(blk.Height)

	//for genesis block
	if blk.Height == 0 && !acc.HasDelegatedStake() {
		acc.DelegatedStakeVersion = 1
		acc.DelegatedStake = new(dpos.DelegatedStake)
		acc.DelegatedStake.DelegatedPublicKey = blk.DelegatedPublicKey
	}

	acc.DelegatedStake.AddStakePendingStake(reward, blk.Height)
	acc.DelegatedStake.AddStakePendingStake(allFees[config.NATIVE_TOKEN_STRING], blk.Height)
	for key, value := range allFees {
		if key != config.NATIVE_TOKEN_STRING {
			acc.AddBalance(true, value, []byte(key))
		}
	}
	acs.UpdateAccount(blk.Forger, acc)

	tok := toks.GetToken(config.NATIVE_TOKEN_FULL)
	tok.AddSupply(true, reward)
	toks.UpdateToken(config.NATIVE_TOKEN_FULL, tok)

}

func (blk *Block) ComputeHash() cryptography.Hash {
	return cryptography.SHA3Hash(blk.Serialize())
}

func (blk *Block) ComputeKernelHashOnly() cryptography.Hash {
	out := blk.serializeBlock(true, false)
	return cryptography.SHA3Hash(out)
}

func (blk *Block) ComputeKernelHash() cryptography.Hash {
	hash := blk.ComputeKernelHashOnly()
	if blk.Height == 0 {
		return hash
	}
	return cryptography.ComputeKernelHash(hash, blk.StakingAmount)
}

func (blk *Block) SerializeForSigning() cryptography.Hash {
	return cryptography.SHA3Hash(blk.serializeBlock(false, false))
}

func (blk *Block) VerifySignature() bool {
	hash := blk.SerializeForSigning()
	return ecdsa.VerifySignature(blk.DelegatedPublicKey[:], hash[:], blk.Signature[0:64])
}

func (blk *Block) serializeBlock(kernelHash bool, inclSignature bool) []byte {

	writer := helpers.NewBufferWriter()

	blk.BlockHeader.Serialize(writer)

	if !kernelHash {
		writer.Write(blk.MerkleHash[:])
		writer.Write(blk.PrevHash[:])
	}

	writer.Write(blk.PrevKernelHash[:])

	if !kernelHash {

		writer.WriteUvarint(blk.StakingAmount)
		writer.Write(blk.DelegatedPublicKey[:])
	}

	writer.WriteUvarint(blk.Timestamp)

	writer.Write(blk.Forger[:])

	if inclSignature {
		writer.Write(blk.Signature[:])
	}

	return writer.Bytes()
}

func (blk *Block) SerializeForForging() []byte {
	return blk.serializeBlock(true, false)
}

func (blk *Block) Serialize() []byte {
	return blk.serializeBlock(false, true)
}

func (blk *Block) Deserialize(reader *helpers.BufferReader) {
	blk.BlockHeader.Deserialize(reader)
	blk.MerkleHash = reader.ReadHash()
	blk.PrevHash = reader.ReadHash()
	blk.PrevKernelHash = reader.ReadHash()
	blk.StakingAmount = reader.ReadUvarint()
	blk.DelegatedPublicKey = reader.Read33()
	blk.Timestamp = reader.ReadUvarint()
	blk.Forger = reader.Read20()
	blk.Signature = reader.Read65()
}
