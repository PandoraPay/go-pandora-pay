package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"math/big"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts/account"
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

	accs, err := dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL)
	if err != nil {
		return
	}

	var tempAcc *account.Account
	if tempAcc, err = accs.GetAccount(newAccountPublicKey); err != nil {
		return
	}

	tempAcc.Balance.AddBalanceUint(payloadExtra.Reward)

	return accs.Update(string(newAccountPublicKey), tempAcc)
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) AfterIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	newAccountPublicKey := publicKeyList[payloadExtra.TemporaryAccountRegistrationIndex]

	accs, err := dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL)
	if err != nil {
		return
	}

	accs.Delete(string(newAccountPublicKey))
	dataStorage.Regs.Delete(string(newAccountPublicKey))

	return
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) ComputeAllKeys(out map[string]bool) {
}

func (payloadExtra *TransactionZetherPayloadExtraStakingReward) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, payloadParity bool) error {

	if bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}
	if payloadBurnValue != 0 {
		return errors.New("Payload burn value must be zero")
	}
	if payloadStatement.Fee != 0 {
		return errors.New("Payload Fee should have been zero")
	}

	if payloadExtra.Reward == 0 {
		return errors.New("Payload reward must be greater than zero")
	}

	if int(payloadExtra.TemporaryAccountRegistrationIndex) >= len(payloadRegistrations.Registrations) {
		return errors.New("RegistrationIndex is invalid")
	}

	if len(payloadStatement.C) < 64 {
		return errors.New("Payload Extra Reward should had 256 ring members")
	}

	for i, registration := range payloadRegistrations.Registrations {
		if (i%2 == 0) == payloadParity { //sender
			if registration != nil && uint64(i) != payloadExtra.TemporaryAccountRegistrationIndex {
				return errors.New("Payload Reward should not have new sender account")
			}
		}
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

	if payloadExtra.TemporaryAccountRegistrationIndex >= uint64(len(payloadStatement.C)) {
		return errors.New("TemporaryAccountRegistrationIndex out of bound")
	}

	balance := crypto.ConstructElGamal(payloadStatement.CLn[payloadExtra.TemporaryAccountRegistrationIndex], payloadStatement.CRn[payloadExtra.TemporaryAccountRegistrationIndex])

	balance = balance.Plus(new(big.Int).SetUint64(payloadExtra.Reward))

	payloadStatement.CLn[payloadExtra.TemporaryAccountRegistrationIndex] = balance.Left
	payloadStatement.CRn[payloadExtra.TemporaryAccountRegistrationIndex] = balance.Right

	return
}
