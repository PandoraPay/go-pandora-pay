package transaction_zether

import (
	"bytes"
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZether struct {
	transaction_base_interface.TransactionBaseInterface
	ChainHeight uint64
	ChainHash   []byte
	Payloads    []*transaction_zether_payload.TransactionZetherPayload
	Bloom       *TransactionZetherBloom
}

/**
Zether requires another verification that the bloomed publicKeys, CL, CR are the same
*/
func (tx *TransactionZether) IncludeTransaction(blockHeight uint64, txHash []byte, dataStorage *data_storage.DataStorage) (err error) {

	if tx.ChainHeight > blockHeight {
		return fmt.Errorf("Zether ChainHeight is invalid %d > %d", tx.ChainHeight, blockHeight)
	}

	for payloadIndex, payload := range tx.Payloads {
		if err = payload.IncludePayload(txHash, byte(payloadIndex), tx.Bloom.PublicKeyLists[payloadIndex], blockHeight, dataStorage); err != nil {
			return
		}
	}

	return
}

func (tx *TransactionZether) ComputeFee() (uint64, error) {

	sum := uint64(0)
	for _, payload := range tx.Payloads {
		if bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
			if err := helpers.SafeUint64Add(&sum, payload.Statement.Fee); err != nil {
				return 0, err
			}
		} else {
			value := payload.Statement.Fee
			if err := helpers.SafeUint64Mul(&value, payload.FeeRateMax); err != nil {
				return 0, err
			}
			if err := helpers.SafeUint64Add(&sum, value); err != nil {
				return 0, err
			}
		}
	}

	return sum, nil
}

func (tx *TransactionZether) ComputeAllKeys(out map[string]bool) {

	for _, payload := range tx.Payloads {
		payload.ComputeAllKeys(out)
	}

	return
}

func (tx *TransactionZether) Validate() (err error) {

	if len(tx.Payloads) == 0 {
		return errors.New("You need at least one payload")
	}

	for payloadIndex, payload := range tx.Payloads {
		if err = payload.Validate(byte(payloadIndex)); err != nil {
			return
		}
	}

	return
}

func (tx *TransactionZether) VerifySignatureManually(txHash []byte) bool {

	assetMap := map[string]int{}
	for _, payload := range tx.Payloads {
		if payload.Proof.Verify(payload.Asset, assetMap[string(payload.Asset)], tx.ChainHash, payload.Statement, txHash, payload.BurnValue) == false {
			return false
		}
		assetMap[string(payload.Asset)] = assetMap[string(payload.Asset)] + 1
	}

	return true
}

func (tx *TransactionZether) SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(tx.ChainHeight)
	w.Write(tx.ChainHash)

	w.WriteByte(byte(len(tx.Payloads)))
	for _, payload := range tx.Payloads {
		payload.Serialize(w, inclSignature)
	}

}

func (tx *TransactionZether) Serialize(w *helpers.BufferWriter) {
	tx.SerializeAdvanced(w, true)
}

func (tx *TransactionZether) Deserialize(r *helpers.BufferReader) (err error) {

	if tx.ChainHeight, err = r.ReadUvarint(); err != nil {
		return
	}

	if tx.ChainHash, err = r.ReadBytes(cryptography.HashSize); err != nil {
		return
	}

	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}

	tx.Payloads = make([]*transaction_zether_payload.TransactionZetherPayload, n)
	for i := byte(0); i < n; i++ {
		payload := &transaction_zether_payload.TransactionZetherPayload{}
		if err = payload.Deserialize(r); err != nil {
			return
		}
		tx.Payloads[i] = payload
	}

	return
}

func (tx *TransactionZether) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
