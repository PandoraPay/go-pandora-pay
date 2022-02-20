package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"math/big"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/dpos"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations/transaction_zether_registration"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraStakingReward struct {
	TransactionZetherPayloadExtraInterface
	Reward                            uint64
	TemporaryAccountRegistrationIndex uint64
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	reg := payloadRegistrations.Registrations[payloadExtra.TemporaryAccountRegistrationIndex]
	if reg == nil || reg.RegistrationType != transaction_zether_registration.NOT_REGISTERED {
		return errors.New("Account must not be registered before! It should be a new one")
	}

	newAccountPublicKey := publicKeyList[payloadExtra.TemporaryAccountRegistrationIndex]

	var tempAcc *plain_account.PlainAccount
	if tempAcc, err = dataStorage.PlainAccs.GetPlainAccount(newAccountPublicKey, blockHeight); err != nil {
		return
	}

	tempAcc.DelegatedStake.Version = dpos.STAKING
	tempAcc.DelegatedStake.Balance.AddBalanceUint(payloadExtra.Reward)

	return dataStorage.PlainAccs.Update(string(newAccountPublicKey), tempAcc)
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) IncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	newAccountPublicKey := publicKeyList[payloadExtra.TemporaryAccountRegistrationIndex]

	dataStorage.PlainAccs.Delete(string(newAccountPublicKey))
	dataStorage.Regs.Delete(string(newAccountPublicKey))

	return
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) ComputeAllKeys(out map[string]bool) {
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error {

	if bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}
	if payloadBurnValue != 0 {
		return errors.New("Payload burn value must be zero")
	}

	if payloadExtra.Reward > 0 {
		return errors.New("Payload reward must be greater than zero")
	}

	if int(payloadExtra.TemporaryAccountRegistrationIndex) >= len(payloadRegistrations.Registrations) {
		return errors.New("RegistrationIndex is invalid")
	}

	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(payloadExtra.Reward)
	w.WriteUvarint(payloadExtra.TemporaryAccountRegistrationIndex)
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadExtra.Reward, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadExtra.TemporaryAccountRegistrationIndex, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) UpdateStatement(payloadStatement *crypto.Statement) (err error) {

	serialized := append(payloadStatement.CLn[payloadExtra.TemporaryAccountRegistrationIndex].EncodeCompressed(), payloadStatement.CRn[payloadExtra.TemporaryAccountRegistrationIndex].EncodeCompressed()...)

	var balance *crypto.ElGamal
	if balance, err = new(crypto.ElGamal).Deserialize(serialized); err != nil {
		return
	}

	balance = balance.Plus(new(big.Int).SetUint64(payloadExtra.Reward))

	payloadStatement.CLn[payloadExtra.TemporaryAccountRegistrationIndex] = balance.Left
	payloadStatement.CRn[payloadExtra.TemporaryAccountRegistrationIndex] = balance.Right

	return
}
