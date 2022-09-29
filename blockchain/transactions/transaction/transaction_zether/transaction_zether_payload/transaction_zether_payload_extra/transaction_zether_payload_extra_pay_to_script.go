package transaction_zether_payload_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraPayInFuture struct {
	TransactionZetherPayloadExtraInterface
	Deadline           uint64
	DefaultResolution  bool //true for receiver, false for refunding sender
	MultisigThreshold  byte
	MultisigPublicKeys [][]byte
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) AfterIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	//to pay for registering accounts
	for _, publicKey := range publicKeyList {
		if _, _, err = dataStorage.GetOrCreateAccount(payloadAsset, publicKey, false); err != nil {
			return
		}
	}
	return
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) ComputeAllKeys(out map[string]bool) {
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) VerifyExtraSignature(hashForSignature []byte, payloadStatement *crypto.Statement) bool {
	return false
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, payloadParity bool) error {
	if payloadExtra.Deadline > 100000 {
		return errors.New("Deadline should be smaller than 100000")
	}
	if payloadExtra.Deadline < 10 {
		return errors.New("Deadline should be greater than 10")
	}
	if payloadBurnValue != 0 {
		return errors.New("Payload burn value must be zero")
	}
	if payloadStatement.Fee != 0 {
		return errors.New("Payload Fee must be zero")
	}
	if len(payloadExtra.MultisigPublicKeys) > 5 {
		return errors.New("PublicKeys list is limited to 5 elements")
	}
	if payloadExtra.MultisigThreshold == 0 {
		return errors.New("Threshold should not be zero")
	}
	if int(payloadExtra.MultisigThreshold) > len(payloadExtra.MultisigPublicKeys) {
		return errors.New("Invalid threshold")
	}
	unique := make(map[string]bool)
	for i := range payloadExtra.MultisigPublicKeys {
		unique[string(payloadExtra.MultisigPublicKeys[i])] = true
	}
	if len(unique) != len(payloadExtra.MultisigPublicKeys) {
		return errors.New("Duplicate Keys detected")
	}
	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(payloadExtra.Deadline)
	w.WriteBool(payloadExtra.DefaultResolution)
	w.WriteByte(payloadExtra.MultisigThreshold)
	w.WriteByte(byte(len(payloadExtra.MultisigPublicKeys)))
	for _, pb := range payloadExtra.MultisigPublicKeys {
		w.Write(pb)
	}
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) Deserialize(r *helpers.BufferReader) (err error) {

	if payloadExtra.Deadline, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadExtra.DefaultResolution, err = r.ReadBool(); err != nil {
		return
	}

	if payloadExtra.MultisigThreshold, err = r.ReadByte(); err != nil {
		return
	}

	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}
	payloadExtra.MultisigPublicKeys = make([][]byte, n)
	for i := 0; i < len(payloadExtra.MultisigPublicKeys); i++ {
		if payloadExtra.MultisigPublicKeys[i], err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
	}

	return
}

func (payloadExtra *TransactionZetherPayloadExtraPayInFuture) UpdateStatement(payloadStatement *crypto.Statement) error {
	return nil
}
