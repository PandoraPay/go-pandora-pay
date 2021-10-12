package transaction_zether_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherDelegateStake struct {
	TransactionZetherExtraInterface
	DelegatePublicKey []byte
	DelegateSignature []byte

	HasNewDelegatedInfo bool
	NewPublicKey        []byte
	NewFee              uint64
}

func (tx *TransactionZetherDelegateStake) IncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, payloads []*transaction_zether_payload.TransactionZetherPayload, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(tx.DelegatePublicKey, blockHeight); err != nil {
		return
	}

	if plainAcc == nil {
		plainAcc = plain_account.NewPlainAccount(tx.DelegatePublicKey)
	}
	if !plainAcc.HasDelegatedStake() {
		if !tx.HasNewDelegatedInfo {
			return errors.New("HasNewDelegatedInfo is set false")
		}
		if err = plainAcc.CreateDelegatedStake(payloads[0].BurnValue, tx.NewPublicKey, tx.NewFee); err != nil {
			return
		}
	} else {
		if err = plainAcc.DelegatedStake.AddStakePendingStake(payloads[0].BurnValue, blockHeight); err != nil {
			return
		}
		if tx.HasNewDelegatedInfo {
			plainAcc.DelegatedStake.DelegatedStakePublicKey = tx.NewPublicKey
			plainAcc.DelegatedStake.DelegatedStakeFee = tx.NewFee
		}
	}

	return nil
}

func (tx *TransactionZetherDelegateStake) Validate(payloads []*transaction_zether_payload.TransactionZetherPayload) error {

	if len(payloads) != 1 {
		return errors.New("Payloads length must be 1")
	}
	if bytes.Equal(payloads[0].Asset, config_coins.NATIVE_ASSET) {
		return errors.New("Payload[0] asset must be a native asset")
	}

	if tx.HasNewDelegatedInfo {
		if len(tx.NewPublicKey) != cryptography.PublicKeySize {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.NewFee > config_stake.DELEGATING_STAKING_FEES_MAX_VALUE {
			return errors.New("Invalid NewFee")
		}
	} else {
		if len(tx.NewPublicKey) != 0 {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.NewFee != 0 {
			return errors.New("Invalid NewFee")
		}
	}

	return nil
}

func (tx *TransactionZetherDelegateStake) VerifySignatureManually(hashForSignature []byte) bool {
	return crypto.VerifySignature(hashForSignature, tx.DelegateSignature, tx.DelegatePublicKey)
}

func (tx *TransactionZetherDelegateStake) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(tx.DelegatePublicKey)
	w.WriteBool(tx.HasNewDelegatedInfo)
	if tx.HasNewDelegatedInfo {
		if inclSignature {
			w.Write(tx.DelegateSignature)
		}
		w.Write(tx.NewPublicKey)
		w.WriteUvarint(tx.NewFee)
	}
}

func (tx *TransactionZetherDelegateStake) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if tx.HasNewDelegatedInfo, err = r.ReadBool(); err != nil {
		return
	}
	if tx.HasNewDelegatedInfo {
		if tx.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
		if tx.NewPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if tx.NewFee, err = r.ReadUvarint(); err != nil {
			return
		}
	}
	return
}
