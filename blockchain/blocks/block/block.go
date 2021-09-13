package block

import (
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/config/config_reward"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type Block struct {
	*BlockHeader
	MerkleHash         helpers.HexBytes `json:"merkleHash"`     //32 byte
	PrevHash           helpers.HexBytes `json:"prevHash"`       //32 byte
	PrevKernelHash     helpers.HexBytes `json:"prevKernelHash"` //32 byte
	Timestamp          uint64           `json:"timestamp"`
	StakingAmount      uint64           `json:"stakingAmount"`
	Forger             helpers.HexBytes `json:"forger"`             // 33 byte public key
	DelegatedPublicKey helpers.HexBytes `json:"delegatedPublicKey"` // 33 byte public key can also be found into the accounts tree
	Signature          helpers.HexBytes `json:"signature"`          // 64 byte signature
	Bloom              *BlockBloom      `json:"bloom"`
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

func (blk *Block) IncludeBlock(accsCollection *accounts.AccountsCollection, toks *tokens.Tokens, allFees uint64) (err error) {

	accs, err := accsCollection.GetMap(config.NATIVE_TOKEN_FULL)
	if err != nil {
		return
	}

	reward := config_reward.GetRewardAt(blk.Height)

	var acc *account.Account
	if acc, err = accs.GetAccount(blk.Forger, blk.Height); err != nil {
		return
	}
	if acc == nil {
		return errors.New("Account not found")
	}

	if err = acc.NativeExtra.DelegatedStake.AddStakePendingStake(reward, blk.Height); err != nil {
		return
	}
	if err = acc.NativeExtra.DelegatedStake.AddStakePendingStake(allFees, blk.Height); err != nil {
		return
	}

	if err = accs.UpdateAccount(blk.Forger, acc); err != nil {
		return
	}

	var tok *token.Token
	if tok, err = toks.GetToken(config.NATIVE_TOKEN_FULL); err != nil {
		return
	}

	if err = tok.AddSupply(true, reward); err != nil {
		return
	}
	if err = toks.UpdateToken(config.NATIVE_TOKEN_FULL, tok); err != nil {
		return
	}

	return
}

func (blk *Block) computeHash() []byte {
	return cryptography.SHA3(blk.SerializeToBytes())
}

func (blk *Block) ComputeKernelHashOnly() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, true, false)
	return cryptography.SHA3(writer.Bytes())
}

func (blk *Block) ComputeKernelHash() []byte {
	hash := blk.ComputeKernelHashOnly()
	return cryptography.ComputeKernelHash(hash, blk.StakingAmount)
}

func (blk *Block) SerializeForSigning() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, false, false)
	return cryptography.SHA3(writer.Bytes())
}

func (blk *Block) VerifySignatureManually() bool {
	hash := blk.SerializeForSigning()
	return crypto.VerifySignature(hash, blk.Signature, blk.DelegatedPublicKey)
}

func (blk *Block) AdvancedSerialization(w *helpers.BufferWriter, kernelHash bool, inclSignature bool) {

	blk.BlockHeader.Serialize(w)

	if !kernelHash {
		w.Write(blk.MerkleHash)
		w.Write(blk.PrevHash)
	}

	w.Write(blk.PrevKernelHash)

	if !kernelHash {
		w.WriteUvarint(blk.StakingAmount)
	}

	w.WriteUvarint(blk.Timestamp)

	w.Write(blk.Forger)
	if !kernelHash {
		w.Write(blk.DelegatedPublicKey)
	}

	if inclSignature {
		w.Write(blk.Signature)
	}
}

func (blk *Block) SerializeForForging(w *helpers.BufferWriter) {
	blk.AdvancedSerialization(w, true, false)
}

func (blk *Block) Serialize(w *helpers.BufferWriter) {
	w.Write(blk.Bloom.Serialized)
}

func (blk *Block) SerializeToBytes() []byte {
	return blk.Bloom.Serialized
}

func (blk *Block) SerializeManualToBytes() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, false, true)
	return writer.Bytes()
}

func (blk *Block) Deserialize(r *helpers.BufferReader) (err error) {

	first := r.Position

	if err = blk.BlockHeader.Deserialize(r); err != nil {
		return
	}
	if blk.MerkleHash, err = r.ReadHash(); err != nil {
		return
	}
	if blk.PrevHash, err = r.ReadHash(); err != nil {
		return
	}
	if blk.PrevKernelHash, err = r.ReadHash(); err != nil {
		return
	}
	if blk.StakingAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if blk.Timestamp, err = r.ReadUvarint(); err != nil {
		return
	}
	if blk.Forger, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if blk.DelegatedPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if blk.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}

	serialized := r.Buf[first:r.Position]
	blk.BloomSerializedNow(serialized)

	return
}
