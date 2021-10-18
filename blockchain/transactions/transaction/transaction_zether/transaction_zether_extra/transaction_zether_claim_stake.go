package transaction_zether_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherClaimStake struct {
	TransactionZetherExtraInterface
	DelegatePublicKey           []byte
	DelegatedStakingClaimAmount uint64
	RegistrationIndex           byte
	DelegateSignature           []byte
}

func (tx *TransactionZetherClaimStake) BeforeIncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, payloads []*transaction_zether_payload.TransactionZetherPayload, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	var acc *account.Account
	var exists bool

	reg := txRegistrations.Registrations[tx.RegistrationIndex]
	publicKey := publicKeyListByCounter[reg.PublicKeyIndex]

	if accs, err = dataStorage.AccsCollection.GetMap(payloads[0].Asset); err != nil {
		return
	}

	if exists, err = accs.Exists(string(publicKey)); err != nil {
		return
	}

	if exists {
		return errors.New("Account should not exist!")
	}
	if acc, err = accs.CreateAccount(publicKey); err != nil {
		return
	}

	if err = acc.Balance.AddBalanceUint(tx.DelegatedStakingClaimAmount); err != nil {
		return
	}

	return accs.Update(string(publicKey), acc)
}

func (tx *TransactionZetherClaimStake) IncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, payloads []*transaction_zether_payload.TransactionZetherPayload, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	if accs, err = dataStorage.AccsCollection.GetMap(payloads[0].Asset); err != nil {
		return
	}

	reg := txRegistrations.Registrations[tx.RegistrationIndex]
	publicKey := publicKeyListByCounter[reg.PublicKeyIndex]

	accs.Delete(string(publicKey))

	return
}

func (tx *TransactionZetherClaimStake) Validate(txRegistrations *transaction_data.TransactionDataTransactions, payloads []*transaction_zether_payload.TransactionZetherPayload) error {

	if len(payloads) != 1 {
		return errors.New("Payloads length must be 1")
	}
	if bytes.Equal(payloads[0].Asset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}
	if int(tx.RegistrationIndex) >= len(txRegistrations.Registrations) {
		return errors.New("RegistrationIndex is invalid")
	}

	if len(tx.DelegatePublicKey) != cryptography.PublicKeySize || len(tx.DelegateSignature) != cryptography.SignatureSize {
		return errors.New("DelegatePublicKey or DelegateSignature length is invalid")
	}
	if tx.DelegatedStakingClaimAmount == 0 {
		return errors.New("ClaimAmount must be > 0")
	}

	return nil
}

func (tx *TransactionZetherClaimStake) VerifySignatureManually(hashForSignature []byte) bool {
	return crypto.VerifySignature(hashForSignature, tx.DelegateSignature, tx.DelegatePublicKey)
}

func (tx *TransactionZetherClaimStake) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(tx.DelegatePublicKey)
	w.WriteUvarint(tx.DelegatedStakingClaimAmount)
	w.WriteByte(tx.RegistrationIndex)
	if inclSignature {
		w.Write(tx.DelegateSignature)
	}
}

func (tx *TransactionZetherClaimStake) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if tx.DelegatedStakingClaimAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if tx.RegistrationIndex, err = r.ReadByte(); err != nil {
		return
	}
	if tx.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
