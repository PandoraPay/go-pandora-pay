package transaction_zether_extra

import (
	"bytes"
	"errors"
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
	DelegateSignature           []byte
}

func (tx *TransactionZetherClaimStake) Validate(payloads []*transaction_zether_payload.TransactionZetherPayload) error {

	if len(payloads) != 1 {
		return errors.New("Payloads length must be 1")
	}
	if bytes.Equal(payloads[0].Asset, config_coins.NATIVE_ASSET) == false {
		return errors.New("Payload[0] asset must be a native asset")
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
	if tx.DelegateSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
