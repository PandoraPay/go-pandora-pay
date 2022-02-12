package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"math/big"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations/transaction_zether_registration"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraClaim struct {
	TransactionZetherPayloadExtraInterface
	DelegatePublicKey           []byte
	DelegatedStakingClaimAmount uint64
	RegistrationIndex           uint64
	DelegateSignature           []byte
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) getAmount(payloadStatement *crypto.Statement) (uint64, error) {
	amount := payloadExtra.DelegatedStakingClaimAmount
	if err := helpers.SafeUint64Add(&amount, payloadStatement.Fee); err != nil {
		return 0, err
	}
	return amount, nil
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	var acc *account.Account

	amount, err := payloadExtra.getAmount(payloadStatement)
	if err != nil {
		return
	}

	plainAcc, err := dataStorage.PlainAccs.GetPlainAccount(payloadExtra.DelegatePublicKey, blockHeight)
	if err != nil {
		return
	}
	if plainAcc == nil {
		return errors.New("PlainAccount doesn't exist")
	}

	if err = dataStorage.SubtractUnclaimed(plainAcc, amount, blockHeight); err != nil {
		return
	}

	if err = dataStorage.PlainAccs.Update(string(plainAcc.PublicKey), plainAcc); err != nil {
		return
	}

	reg := payloadRegistrations.Registrations[payloadExtra.RegistrationIndex]
	if reg != nil && reg.RegistrationType != transaction_zether_registration.NOT_REGISTERED {
		return errors.New("Account must not be registered before! It should be a new one")
	}

	publicKey := publicKeyList[payloadExtra.RegistrationIndex]

	if accs, err = dataStorage.AccsCollection.GetMap(payloadAsset); err != nil {
		return
	}

	if acc, err = accs.GetAccount(publicKey); err != nil {
		return
	}

	acc.Balance.AddBalanceUint(amount)

	return accs.Update(string(publicKey), acc)
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) IncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	if accs, err = dataStorage.AccsCollection.GetMap(payloadAsset); err != nil {
		return
	}

	publicKey := publicKeyList[payloadExtra.RegistrationIndex]

	accs.Delete(string(publicKey))
	dataStorage.Regs.Delete(string(publicKey))

	return
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) ComputeAllKeys(out map[string]bool) {
	out[string(payloadExtra.DelegatePublicKey)] = true
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error {

	if bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}
	if payloadBurnValue != 0 {
		return errors.New("Payload burn value must be zero")
	}

	if int(payloadExtra.RegistrationIndex) >= len(payloadRegistrations.Registrations) {
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

func (payloadExtra *TransactionZetherPayloadExtraClaim) VerifyExtraSignature(hashForSignature []byte) bool {
	return crypto.VerifySignature(hashForSignature, payloadExtra.DelegateSignature, payloadExtra.DelegatePublicKey)
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(payloadExtra.DelegatePublicKey)
	w.WriteUvarint(payloadExtra.DelegatedStakingClaimAmount)
	w.WriteUvarint(payloadExtra.RegistrationIndex)
	if inclSignature {
		w.Write(payloadExtra.DelegateSignature)
	}
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadExtra.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if payloadExtra.DelegatedStakingClaimAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadExtra.RegistrationIndex, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadExtra.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}

	return
}

func (payloadExtra *TransactionZetherPayloadExtraClaim) UpdateStatement(payloadStatement *crypto.Statement) (err error) {

	serialized := append(payloadStatement.CLn[payloadExtra.RegistrationIndex].EncodeCompressed(), payloadStatement.CRn[payloadExtra.RegistrationIndex].EncodeCompressed()...)

	var balance *crypto.ElGamal
	if balance, err = new(crypto.ElGamal).Deserialize(serialized); err != nil {
		return
	}

	amount, err := payloadExtra.getAmount(payloadStatement)
	if err != nil {
		return
	}

	balance = balance.Plus(new(big.Int).SetUint64(amount))

	payloadStatement.CLn[payloadExtra.RegistrationIndex] = balance.Left
	payloadStatement.CRn[payloadExtra.RegistrationIndex] = balance.Right

	return
}
