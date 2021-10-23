package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraClaimStake struct {
	TransactionZetherPayloadExtraInterface
	DelegatePublicKey           []byte
	DelegatedStakingClaimAmount uint64
	RegistrationIndex           byte
	DelegateSignature           []byte
}

func (payloadExtra *TransactionZetherPayloadExtraClaimStake) BeforeIncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex int, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	var acc *account.Account
	var exists bool

	amount := payloadExtra.DelegatedStakingClaimAmount
	if err = helpers.SafeUint64Add(&amount, payloadStatement.Fees); err != nil {
		return
	}

	plainAcc, err := dataStorage.PlainAccs.GetPlainAccount(payloadExtra.DelegatePublicKey, blockHeight)
	if err != nil {
		return
	}
	if plainAcc == nil {
		return errors.New("PlainAccount doesn't exist")
	}

	if err = plainAcc.AddUnclaimed(false, amount); err != nil {
		return
	}

	if err = dataStorage.PlainAccs.Update(string(plainAcc.PublicKey), plainAcc); err != nil {
		return
	}

	reg := txRegistrations.Registrations[payloadExtra.RegistrationIndex]
	publicKey := publicKeyList[reg.PublicKeyIndex]

	if accs, err = dataStorage.AccsCollection.GetMap(payloadAsset); err != nil {
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

	if err = acc.Balance.AddBalanceUint(amount); err != nil {
		return
	}

	return accs.Update(string(publicKey), acc)
}

func (payloadExtra *TransactionZetherPayloadExtraClaimStake) IncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex int, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	if accs, err = dataStorage.AccsCollection.GetMap(payloadAsset); err != nil {
		return
	}

	reg := txRegistrations.Registrations[payloadExtra.RegistrationIndex]
	publicKey := publicKeyList[reg.PublicKeyIndex]

	accs.Delete(string(publicKey))

	return
}

func (payloadExtra *TransactionZetherPayloadExtraClaimStake) Validate(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex int, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error {

	if bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}
	if payloadBurnValue != 0 {
		return errors.New("Payload burn value must be zero")
	}

	if int(payloadExtra.RegistrationIndex) >= len(txRegistrations.Registrations) {
		return errors.New("RegistrationIndex is invalid")
	}

	if len(payloadExtra.DelegatePublicKey) != cryptography.PublicKeySize || len(payloadExtra.DelegateSignature) != cryptography.SignatureSize {
		return errors.New("DelegatePublicKey or DelegateSignature length is invalid")
	}
	if payloadExtra.DelegatedStakingClaimAmount == 0 {
		return errors.New("ClaimAmount must be > 0")
	}

	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraClaimStake) VerifyExtraSignature(hashForSignature []byte) bool {
	return crypto.VerifySignature(hashForSignature, payloadExtra.DelegateSignature, payloadExtra.DelegatePublicKey)
}

func (payloadExtra *TransactionZetherPayloadExtraClaimStake) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(payloadExtra.DelegatePublicKey)
	w.WriteUvarint(payloadExtra.DelegatedStakingClaimAmount)
	w.WriteByte(payloadExtra.RegistrationIndex)
	if inclSignature {
		w.Write(payloadExtra.DelegateSignature)
	}
}

func (payloadExtra *TransactionZetherPayloadExtraClaimStake) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadExtra.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if payloadExtra.DelegatedStakingClaimAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadExtra.RegistrationIndex, err = r.ReadByte(); err != nil {
		return
	}
	if payloadExtra.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
