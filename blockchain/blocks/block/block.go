package block

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/config/config_reward"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type Block struct {
	*BlockHeader
	MerkleHash     helpers.HexBytes `json:"merkleHash"`     //32 byte
	PrevHash       helpers.HexBytes `json:"prevHash"`       //32 byte
	PrevKernelHash helpers.HexBytes `json:"prevKernelHash"` //32 byte
	Timestamp      uint64           `json:"timestamp"`
	StakingAmount  uint64           `json:"stakingAmount"`
	Forger         helpers.HexBytes `json:"forger"`    // 20 byte public key hash
	Signature      helpers.HexBytes `json:"signature"` // 65 byte signature
	Bloom          *BlockBloom      `json:"bloom"`
}

func CreateEmptyBlock() *Block {
	return &Block{
		BlockHeader: &BlockHeader{},
	}
}

func (blk *Block) Validate() error {
	return blk.BlockHeader.Validate()
}

func (blk *Block) Verify() error {
	return blk.Bloom.verifyIfBloomed()
}

func (blk *Block) IncludeBlock(acs *accounts.Accounts, toks *tokens.Tokens, allFees map[string]uint64) (err error) {

	reward := config_reward.GetRewardAt(blk.Height)

	var acc *account.Account
	if acc, err = acs.GetAccountEvenEmpty(blk.Forger, blk.Height); err != nil {
		return
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
	if err = acs.UpdateAccount(blk.Forger, acc); err != nil {
		return
	}

	var tok *token.Token
	if tok, err = toks.GetToken(config.NATIVE_TOKEN); err != nil {
		return
	}

	if err = tok.AddSupply(true, reward); err != nil {
		return
	}
	if err = toks.UpdateToken(config.NATIVE_TOKEN, tok); err != nil {
		return
	}
	return
}

func (blk *Block) computeHash() []byte {
	return cryptography.SHA3Hash(blk.SerializeToBytes())
}

func (blk *Block) ComputeKernelHashOnly() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, true, false)
	return cryptography.SHA3Hash(writer.Bytes())
}

func (blk *Block) ComputeKernelHash() []byte {
	hash := blk.ComputeKernelHashOnly()
	return cryptography.ComputeKernelHash(hash, blk.StakingAmount)
}

func (blk *Block) SerializeForSigning() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, false, false)
	return cryptography.SHA3Hash(writer.Bytes())
}

func (blk *Block) VerifySignatureManually() bool {
	hash := blk.SerializeForSigning()
	publicKey, err := ecdsa.EcrecoverCompressed(hash, blk.Signature)
	if err != nil {
		return false
	}
	return ecdsa.VerifySignature(publicKey, hash, blk.Signature[0:64])
}

func (blk *Block) AdvancedSerialization(writer *helpers.BufferWriter, kernelHash bool, inclSignature bool) {

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
}

func (blk *Block) SerializeForForging(writer *helpers.BufferWriter) {
	blk.AdvancedSerialization(writer, true, false)
}

func (blk *Block) Serialize(writer *helpers.BufferWriter) {
	writer.Write(blk.Bloom.Serialized)
}

func (blk *Block) SerializeToBytes() []byte {
	return blk.Bloom.Serialized
}

func (blk *Block) SerializeManualToBytes() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, false, true)
	return writer.Bytes()
}

func (blk *Block) Deserialize(reader *helpers.BufferReader) (err error) {

	first := reader.Position

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
	if blk.Forger, err = reader.ReadBytes(cryptography.PublicKeyHashHashSize); err != nil {
		return
	}
	if blk.Signature, err = reader.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}

	serialized := reader.Buf[first:reader.Position]
	blk.BloomSerializedNow(serialized)

	return
}
