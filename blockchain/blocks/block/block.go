package block

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_reward"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Block struct {
	*BlockHeader
	MerkleHash     []byte      `json:"merkleHash" msgpack:"merkleHash"`          //32 byte
	PrevHash       []byte      `json:"prevHash"  msgpack:"prevHash"`             //32 byte
	PrevKernelHash []byte      `json:"prevKernelHash"  msgpack:"prevKernelHash"` //32 byte
	Timestamp      uint64      `json:"timestamp" msgpack:"timestamp"`
	StakingAmount  uint64      `json:"stakingAmount" msgpack:"stakingAmount"`
	StakingNonce   []byte      `json:"stakingNonce" msgpack:"stakingNonce"` // 33 byte public key can also be found into the accounts tree
	StakingFee     uint64      `json:"stakingFee" msgpack:"stakingFee"`
	Bloom          *BlockBloom `json:"bloom" msgpack:"bloom"`
}

func CreateEmptyBlock() *Block {
	return &Block{
		BlockHeader: &BlockHeader{},
	}
}

func (blk *Block) validate() error {
	if err := blk.BlockHeader.validate(); err != nil {
		return err
	}

	if blk.StakingFee > config_stake.DELEGATING_STAKING_FEE_MAX_VALUE {
		return errors.New("DelegatedStakeFee is invalid")
	}

	return nil
}

func (blk *Block) Verify() error {
	return blk.Bloom.verifyIfBloomed()
}

func (blk *Block) IncludeBlock(dataStorage *data_storage.DataStorage, allFees uint64) (err error) {

	reward := config_reward.GetRewardAt(blk.Height)

	var ast *asset.Asset
	if ast, err = dataStorage.Asts.GetAsset(config_coins.NATIVE_ASSET_FULL); err != nil {
		return
	}

	if err = ast.AddNativeSupply(true, reward); err != nil {
		return
	}
	if err = dataStorage.Asts.Update(string(config_coins.NATIVE_ASSET_FULL), ast); err != nil {
		return
	}

	return
}

func (blk *Block) computeHash() []byte {
	return cryptography.SHA3(helpers.SerializeToBytes(blk))
}

func (blk *Block) ComputeKernelHashOnly() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, true, false)
	return cryptography.SHA3(writer.Bytes())
}

func (blk *Block) ComputeKernelHash() ([]byte, error) {
	hash := blk.ComputeKernelHashOnly()
	return cryptography.ComputeKernelHash(hash, blk.StakingAmount)
}

func (blk *Block) SerializeForSigning() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, false, false)
	return cryptography.SHA3(writer.Bytes())
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

	w.Write(blk.StakingNonce)

	if !kernelHash {
		w.WriteUvarint(blk.StakingFee)
	}

}

func (blk *Block) SerializeForForging(w *helpers.BufferWriter) {
	blk.AdvancedSerialization(w, true, false)
}

func (blk *Block) Serialize(w *helpers.BufferWriter) {
	w.Write(blk.Bloom.Serialized)
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
	if blk.StakingNonce, err = r.ReadBytes(33); err != nil {
		return
	}
	if blk.StakingFee, err = r.ReadUvarint(); err != nil {
		return
	}

	serialized := r.Buf[first:r.Position]
	blk.BloomSerializedNow(serialized)

	return
}
