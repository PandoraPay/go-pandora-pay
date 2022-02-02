package block

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_reward"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type Block struct {
	*BlockHeader
	MerkleHash               helpers.HexBytes `json:"merkleHash" msgpack:"merkleHash"`          //32 byte
	PrevHash                 helpers.HexBytes `json:"prevHash"  msgpack:"prevHash"`             //32 byte
	PrevKernelHash           helpers.HexBytes `json:"prevKernelHash"  msgpack:"prevKernelHash"` //32 byte
	Timestamp                uint64           `json:"timestamp" msgpack:"timestamp"`
	StakingAmount            uint64           `json:"stakingAmount" msgpack:"stakingAmount"`
	Forger                   helpers.HexBytes `json:"forger" msgpack:"forger"`                                   // 33 byte public key
	DelegatedStakePublicKey  helpers.HexBytes `json:"delegatedStakePublicKey" msgpack:"delegatedStakePublicKey"` // 33 byte public key can also be found into the accounts tree
	DelegatedStakeFee        uint64           `json:"delegatedStakeFee" msgpack:"delegatedStakeFee"`
	RewardCollectorPublicKey helpers.HexBytes `json:"rewardCollectorPublicKey" msgpack:"rewardCollectorPublicKey"` // 33 byte public key only if rewardFee > 0
	Signature                helpers.HexBytes `json:"signature" msgpack:"signature"`                               // 64 byte signature
	Bloom                    *BlockBloom      `json:"bloom" msgpack:"bloom"`
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

	if blk.DelegatedStakeFee == 0 && len(blk.RewardCollectorPublicKey) != 0 {
		return errors.New("blk.RewardCollectorPublicKey must be nil")
	}
	if blk.DelegatedStakeFee > 0 && len(blk.RewardCollectorPublicKey) != cryptography.PublicKeySize {
		return errors.New("blk.RewardCollectorPublicKey invalid length")
	}
	if blk.DelegatedStakeFee > config_stake.DELEGATING_STAKING_FEE_MAX_VALUE {
		return errors.New("DelegatedStakeFee is invalid")
	}
	if bytes.Equal(blk.RewardCollectorPublicKey, blk.DelegatedStakePublicKey) {
		return errors.New("RewardCollectorPublicKey should not be the same with DelegatedStakePublicKey")
	}

	return nil
}

func (blk *Block) Verify() error {
	return blk.Bloom.verifyIfBloomed()
}

func (blk *Block) IncludeBlock(dataStorage *data_storage.DataStorage, allFees uint64) (err error) {

	reward := config_reward.GetRewardAt(blk.Height)

	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(blk.Forger, blk.Height); err != nil {
		return
	}
	if plainAcc == nil {
		return errors.New("Plain Account not found")
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
		var plainAccRewardCollector *plain_account.PlainAccount
		if plainAccRewardCollector, err = dataStorage.GetOrCreatePlainAccount(blk.RewardCollectorPublicKey, blk.Height); err != nil {
			return
		}

		if plainAccRewardCollector.DelegatedStake.HasDelegatedStake() {
			if err = plainAccRewardCollector.DelegatedStake.AddStakePendingStake(commission, blk.Height); err != nil {
				return
			}
		} else {
			if err = plainAccRewardCollector.AddUnclaimed(true, commission); err != nil {
				return
			}
		}

		if err = dataStorage.PlainAccs.Update(string(blk.RewardCollectorPublicKey), plainAccRewardCollector); err != nil {
			return
		}

	}

	if err = plainAcc.DelegatedStake.AddStakePendingStake(final, blk.Height); err != nil {
		return
	}

	if err = dataStorage.PlainAccs.Update(string(blk.Forger), plainAcc); err != nil {
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
	return crypto.VerifySignature(hash, blk.Signature, blk.DelegatedStakePublicKey)
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
			w.Write(blk.RewardCollectorPublicKey)
		}
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
	if blk.DelegatedStakePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if blk.DelegatedStakeFee, err = r.ReadUvarint(); err != nil {
		return
	}
	if blk.DelegatedStakeFee > 0 {
		if blk.RewardCollectorPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
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
