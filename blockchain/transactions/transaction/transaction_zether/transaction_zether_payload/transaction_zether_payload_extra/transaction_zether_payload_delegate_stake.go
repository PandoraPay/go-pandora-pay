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

type TransactionZetherPayloadDelegateStake struct {
	TransactionZetherPayloadExtraInterface
	DelegatePublicKey      []byte
	DelegatedStakingUpdate *transaction_data.TransactionDataDelegatedStakingUpdate
	DelegateSignature      []byte //if newInfo then the signature is required to verify that he is owner
}

func (tx *TransactionZetherPayloadDelegateStake) BeforeIncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) error {
	return nil
}

func (tx *TransactionZetherPayloadDelegateStake) IncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(tx.DelegatePublicKey, blockHeight); err != nil {
		return
	}

	if plainAcc == nil {
		plainAcc = plain_account.NewPlainAccount(tx.DelegatePublicKey)
	}

	if err = tx.DelegatedStakingUpdate.Include(plainAcc); err != nil {
		return
	}

	if err = plainAcc.DelegatedStake.AddStakePendingStake(payloadBurnValue, blockHeight); err != nil {
		return
	}

	if err = dataStorage.PlainAccs.Update(string(tx.DelegatePublicKey), plainAcc); err != nil {
		return
	}

	return nil
}

func (tx *TransactionZetherPayloadDelegateStake) Validate(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) (err error) {

	if bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}
	if payloadBurnValue == 0 {
		return errors.New("Payload burn value must be greater than zero")
	}

	if err = tx.DelegatedStakingUpdate.Validate(); err != nil {
		return
	}

	if tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo && len(tx.DelegateSignature) != cryptography.SignatureSize {
		return errors.New("tx.DelegateSignature length is invalid")
	} else if !tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo && len(tx.DelegateSignature) != 0 {
		return errors.New("tx.DelegateSignature length is not zero")
	}

	return
}

func (tx *TransactionZetherPayloadDelegateStake) VerifyExtraSignature(hashForSignature []byte) bool {
	if tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		return crypto.VerifySignature(hashForSignature, tx.DelegateSignature, tx.DelegatePublicKey)
	}
	return true
}

func (tx *TransactionZetherPayloadDelegateStake) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(tx.DelegatePublicKey)
	tx.DelegatedStakingUpdate.Serialize(w)
	if tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo && inclSignature {
		w.Write(tx.DelegateSignature)
	}
}

func (tx *TransactionZetherPayloadDelegateStake) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	tx.DelegatedStakingUpdate = &transaction_data.TransactionDataDelegatedStakingUpdate{}
	if err = tx.DelegatedStakingUpdate.Deserialize(r); err != nil {
		return
	}
	if tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		if tx.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
	}
	return
}
