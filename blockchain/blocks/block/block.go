package block

import (
	"bytes"
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_reward"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Block struct {
	*BlockHeader
	MerkleHash              []byte      `json:"merkleHash" msgpack:"merkleHash"`          //32 byte
	PrevHash                []byte      `json:"prevHash"  msgpack:"prevHash"`             //32 byte
	PrevKernelHash          []byte      `json:"prevKernelHash"  msgpack:"prevKernelHash"` //32 byte
	Timestamp               uint64      `json:"timestamp" msgpack:"timestamp"`
	StakingAmount           uint64      `json:"stakingAmount" msgpack:"stakingAmount"`
	Forger                  []byte      `json:"forger" msgpack:"forger"`                                   // 20 byte public key
	DelegatedStakePublicKey []byte      `json:"delegatedStakePublicKey" msgpack:"delegatedStakePublicKey"` // 33 byte public key can also be found into the accounts tree
	DelegatedStakeFee       uint64      `json:"delegatedStakeFee" msgpack:"delegatedStakeFee"`
	RewardCollector         []byte      `json:"rewardCollector" msgpack:"rewardCollector"` // 20 byte public key only if rewardFee > 0
	Signature               []byte      `json:"signature" msgpack:"signature"`             // 64 byte signature
	Bloom                   *BlockBloom `json:"bloom" msgpack:"bloom"`
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

	if len(blk.Forger) != cryptography.PublicKeyHashSize {
		return errors.New("Forger is invalid")
	}
	if len(blk.DelegatedStakePublicKey) != cryptography.PublicKeySize {
		return errors.New("DelegatedStakePublicKey is invalid")
	}
	if blk.DelegatedStakeFee == 0 && len(blk.RewardCollector) != 0 {
		return errors.New("blk.RewardCollector must be nil")
	}
	if blk.DelegatedStakeFee > 0 && len(blk.RewardCollector) != cryptography.PublicKeyHashSize {
		return errors.New("blk.RewardCollector invalid length")
	}
	if blk.DelegatedStakeFee > config_stake.DELEGATING_STAKING_FEE_MAX_VALUE {
		return errors.New("DelegatedStakeFee is invalid")
	}
	if bytes.Equal(blk.RewardCollector, blk.DelegatedStakePublicKey) {
		return errors.New("RewardCollector should not be the same with DelegatedStakePublicKey")
	}

	return nil
}

func (blk *Block) Verify() error {
	return blk.Bloom.verifyIfBloomed()
}

func (blk *Block) IncludeBlock(dataStorage *data_storage.DataStorage, allFees uint64) (err error) {

	if blk.StakingAmount < config_stake.GetRequiredStake(blk.Height) {
		return errors.New("Stake amount is not enought")
	}

	reward := config_reward.GetRewardAt(blk.Height)

	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(blk.Forger); err != nil {
		return
	}
	if plainAcc == nil {
		return errors.New("Plain Account not found")
	}

	if blk.StakingAmount != plainAcc.StakeAvailable {
		return fmt.Errorf("Block Staking Amount doesn't match %d %d", blk.StakingAmount, plainAcc.StakeAvailable)
	}

	if blk.DelegatedStakeFee != plainAcc.DelegatedStake.DelegatedStakeFee {
		return fmt.Errorf("Block Delegated Stake Fee doesn't match %d %d", blk.DelegatedStakeFee, plainAcc.DelegatedStake.DelegatedStakeFee)
	}

	final := reward
	if err = helpers.SafeUint64Add(&final, allFees); err != nil {
		return
	}

	if blk.DelegatedStakeFee > 0 {

		//compute the commission
		commission := final
		if err = helpers.SafeUint64Mul(&commission, blk.DelegatedStakeFee); err != nil {
			return
		}
		commission /= config_stake.DELEGATING_STAKING_FEE_MAX_VALUE

		if err = helpers.SafeUint64Sub(&final, commission); err != nil {
			return
		}

		//let's add the commission
		if err = dataStorage.AddStakePendingStake(blk.RewardCollector, commission, true, blk.Height); err != nil {
			return
		}

	}

	if err = dataStorage.AddStakePendingStake(blk.Forger, final, true, blk.Height); err != nil {
		return
	}

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

func (blk *Block) ComputeKernelHash() []byte {
	writer := helpers.NewBufferWriter()
	blk.AdvancedSerialization(writer, true, false)
	return cryptography.SHA3(writer.Bytes())
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

	w.Write(blk.Forger)

	if !kernelHash {
		w.Write(blk.DelegatedStakePublicKey)
		w.WriteUvarint(blk.DelegatedStakeFee)
		if blk.DelegatedStakeFee > 0 {
			w.Write(blk.RewardCollector)
		}
	}

	if !kernelHash && inclSignature {
		w.Write(blk.Signature)
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
	if blk.Forger, err = r.ReadBytes(cryptography.PublicKeyHashSize); err != nil {
		return
	}
	if blk.DelegatedStakePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if blk.DelegatedStakeFee, err = r.ReadUvarint(); err != nil {
		return
	}
	if blk.DelegatedStakeFee > 0 {
		if blk.RewardCollector, err = r.ReadBytes(cryptography.PublicKeyHashSize); err != nil {
			return
		}
	}
	if blk.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}

	serialized := r.Buf[first:r.Position]
	blk.BloomSerializedNow(serialized)

	return
}
