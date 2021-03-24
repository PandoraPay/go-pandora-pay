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
	MerkleHash     helpers.ByteString //32 byte
	PrevHash       helpers.ByteString //32 byte
	PrevKernelHash helpers.ByteString //32 byte
	Timestamp      uint64
	StakingAmount  uint64
	Forger         helpers.ByteString // 20 byte public key hash
	Signature      helpers.ByteString // 65 byte signature
	Bloom          *BlockBloom
}

func (blk *Block) Validate() error {
	return blk.BlockHeader.Validate()
}

func (blk *Block) Verify() error {
	return blk.VerifyBloomAll()
}

func (blk *Block) IncludeBlock(acs *accounts.Accounts, toks *tokens.Tokens, allFees map[string]uint64) (err error) {

	reward := reward.GetRewardAt(blk.Height)
	acc := acs.GetAccountEvenEmpty(blk.Forger)
	if err = acc.RefreshDelegatedStake(blk.Height); err != nil {
		return
	}

	//for genesis block
	if blk.Height == 0 && !acc.HasDelegatedStake() {
		acc.DelegatedStakeVersion = 1
		acc.DelegatedStake = new(dpos.DelegatedStake)
		acc.DelegatedStake.DelegatedPublicKeyHash = blk.Bloom.DelegatedPublicKeyHash
	}

	if err = acc.DelegatedStake.AddStakePendingStake(reward, blk.Height); err != nil {
		return
	}
	if err = acc.DelegatedStake.AddStakePendingStake(allFees[config.NATIVE_TOKEN_STRING], blk.Height); err != nil {
		return
	}
	for key, value := range allFees {
		if key != config.NATIVE_TOKEN_STRING {
			if err = acc.AddBalance(true, value, []byte(key)); err != nil {
				return
			}
		}
	}
	acs.UpdateAccount(blk.Forger, acc)

	tok := toks.GetToken(config.NATIVE_TOKEN)
	if err = tok.AddSupply(true, reward); err != nil {
		return
	}
	toks.UpdateToken(config.NATIVE_TOKEN, tok)
	return
}

func (blk *Block) ComputeHash() []byte {
	return cryptography.SHA3Hash(blk.Serialize())
}

func (blk *Block) ComputeKernelHashOnly() []byte {
	out := blk.serializeBlock(true, false)
	return cryptography.SHA3Hash(out)
}

func (blk *Block) ComputeKernelHash() []byte {
	hash := blk.ComputeKernelHashOnly()
	if blk.Height == 0 {
		return hash
	}
	return cryptography.ComputeKernelHash(hash, blk.StakingAmount)
}

func (blk *Block) SerializeForSigning() []byte {
	return cryptography.SHA3Hash(blk.serializeBlock(false, false))
}

func (blk *Block) VerifySignatureManually() bool {
	hash := blk.SerializeForSigning()
	publicKey, err := ecdsa.EcrecoverCompressed(hash, blk.Signature)
	if err != nil {
		return false
	}
	return ecdsa.VerifySignature(publicKey, hash, blk.Signature[0:64])
}

func (blk *Block) serializeBlock(kernelHash bool, inclSignature bool) []byte {

	writer := helpers.NewBufferWriter()

	blk.BlockHeader.Serialize(writer)

	if !kernelHash {
		writer.Write(blk.MerkleHash)
		writer.Write(blk.PrevHash)
	}

	writer.Write(blk.PrevKernelHash)

	if !kernelHash {
		writer.WriteUvarint(blk.StakingAmount)
	}

	writer.WriteUvarint(blk.Timestamp)

	writer.Write(blk.Forger)

	if inclSignature {
		writer.Write(blk.Signature)
	}

	return writer.Bytes()
}

func (blk *Block) SerializeForForging() []byte {
	return blk.serializeBlock(true, false)
}

func (blk *Block) Serialize() []byte {
	return blk.serializeBlock(false, true)
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
	if blk.Timestamp, err = reader.ReadUvarint(); err != nil {
		return
	}
	if blk.Forger, err = reader.ReadBytes(20); err != nil {
		return
	}
	if blk.Signature, err = reader.ReadBytes(65); err != nil {
		return
	}
	return
}

func (blk *Block) Size() uint64 {
	return uint64(len(blk.Serialize()))
}
