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

	DelegatedStakingNewInfo      bool
	DelegatedStakingNewPublicKey []byte
	DelegatedStakingNewFee       uint64
	DelegateSignature            []byte //if newInfo then the signature is required to verify that he is owner
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
		if !tx.DelegatedStakingNewInfo {
			return errors.New("DelegatedStakingNewInfo is set false")
		}
		if err = plainAcc.CreateDelegatedStake(payloads[0].BurnValue, tx.DelegatedStakingNewPublicKey, tx.DelegatedStakingNewFee); err != nil {
			return
		}
	} else {
		if err = plainAcc.DelegatedStake.AddStakePendingStake(payloads[0].BurnValue, blockHeight); err != nil {
			return
		}
		if tx.DelegatedStakingNewInfo {
			plainAcc.DelegatedStake.DelegatedStakePublicKey = tx.DelegatedStakingNewPublicKey
			plainAcc.DelegatedStake.DelegatedStakeFee = tx.DelegatedStakingNewFee
		}
	}

	return nil
}

func (tx *TransactionZetherDelegateStake) Validate(payloads []*transaction_zether_payload.TransactionZetherPayload) error {

	if len(payloads) != 1 {
		return errors.New("Payloads length must be 1")
	}
	if bytes.Equal(payloads[0].Asset, config_coins.NATIVE_ASSET) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}

	if tx.DelegatedStakingNewInfo {
		if len(tx.DelegatedStakingNewPublicKey) != cryptography.PublicKeySize {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.DelegatedStakingNewFee > config_stake.DELEGATING_STAKING_FEES_MAX_VALUE {
			return errors.New("Invalid NewFee")
		}
	} else {
		if len(tx.DelegatedStakingNewPublicKey) != 0 {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.DelegatedStakingNewFee != 0 {
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
	w.WriteBool(tx.DelegatedStakingNewInfo)
	if tx.DelegatedStakingNewInfo {
		w.Write(tx.DelegatedStakingNewPublicKey)
		w.WriteUvarint(tx.DelegatedStakingNewFee)
		if inclSignature {
			w.Write(tx.DelegateSignature)
		}
	}
}

func (tx *TransactionZetherDelegateStake) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if tx.DelegatedStakingNewInfo, err = r.ReadBool(); err != nil {
		return
	}
	if tx.DelegatedStakingNewInfo {
		if tx.DelegatedStakingNewPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if tx.DelegatedStakingNewFee, err = r.ReadUvarint(); err != nil {
			return
		}
		if tx.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
	}
	return
}
