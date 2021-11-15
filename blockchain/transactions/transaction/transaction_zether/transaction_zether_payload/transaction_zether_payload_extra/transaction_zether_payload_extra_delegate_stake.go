package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraDelegateStake struct {
	TransactionZetherPayloadExtraInterface
	DelegatePublicKey      []byte
	ConvertToUnclaimed     bool
	DelegatedStakingUpdate *transaction_data.TransactionDataDelegatedStakingUpdate
	DelegateSignature      []byte //if newInfo then the signature is required to verify that he is owner
}

func (payloadExtra *TransactionZetherPayloadExtraDelegateStake) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) error {
	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraDelegateStake) IncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(payloadExtra.DelegatePublicKey, blockHeight); err != nil {
		return
	}

	if plainAcc == nil {
		if plainAcc, err = dataStorage.PlainAccs.CreatePlainAccount(payloadExtra.DelegatePublicKey); err != nil {
			return
		}
	}

	if err = payloadExtra.DelegatedStakingUpdate.Include(plainAcc); err != nil {
		return
	}

	if payloadExtra.ConvertToUnclaimed {
		if err = plainAcc.AddUnclaimed(true, payloadBurnValue); err != nil {
			return
		}
	} else {
		if err = plainAcc.DelegatedStake.AddStakePendingStake(payloadBurnValue, blockHeight); err != nil {
			return
		}
	}

	if err = dataStorage.PlainAccs.Update(string(payloadExtra.DelegatePublicKey), plainAcc); err != nil {
		return
	}

	return
}

func (payloadExtra *TransactionZetherPayloadExtraDelegateStake) ComputeAllKeys(out map[string]bool) {
	out[string(payloadExtra.DelegatePublicKey)] = true
}

func (payloadExtra *TransactionZetherPayloadExtraDelegateStake) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) (err error) {

	if bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}
	if payloadBurnValue == 0 {
		return errors.New("Payload burn value must be greater than zero")
	}

	if err = payloadExtra.DelegatedStakingUpdate.Validate(); err != nil {
		return
	}

	if payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo && len(payloadExtra.DelegateSignature) != cryptography.SignatureSize {
		return errors.New("tx.DelegateSignature length is invalid")
	} else if !payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo && len(payloadExtra.DelegateSignature) != 0 {
		return errors.New("tx.DelegateSignature length is not zero")
	}

	return
}

func (payloadExtra *TransactionZetherPayloadExtraDelegateStake) VerifyExtraSignature(hashForSignature []byte) bool {
	if payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		return crypto.VerifySignature(hashForSignature, payloadExtra.DelegateSignature, payloadExtra.DelegatePublicKey)
	}
	return true
}

func (payloadExtra *TransactionZetherPayloadExtraDelegateStake) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(payloadExtra.DelegatePublicKey)
	w.WriteBool(payloadExtra.ConvertToUnclaimed)
	payloadExtra.DelegatedStakingUpdate.Serialize(w)
	if payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo && inclSignature {
		w.Write(payloadExtra.DelegateSignature)
	}
}

func (payloadExtra *TransactionZetherPayloadExtraDelegateStake) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadExtra.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if payloadExtra.ConvertToUnclaimed, err = r.ReadBool(); err != nil {
		return
	}
	payloadExtra.DelegatedStakingUpdate = &transaction_data.TransactionDataDelegatedStakingUpdate{}
	if err = payloadExtra.DelegatedStakingUpdate.Deserialize(r); err != nil {
		return
	}
	if payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		if payloadExtra.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
	}
	return
}
