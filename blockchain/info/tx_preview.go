package info

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
	"pandora-pay/cryptography/crypto"
)

type TxPreview struct {
	TxBase  interface{}                         `json:"base"  msgpack:"base"`
	Version transaction_type.TransactionVersion `json:"version"  msgpack:"version"`
	Hash    []byte                              `json:"hash"  msgpack:"hash"`
	Fee     uint64                              `json:"fee"  msgpack:"fee"`
}

func CreateTxPreviewFromTx(tx *transaction.Transaction) (*TxPreview, error) {

	var base any

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		var dataPublic []byte
		if txBase.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
			dataPublic = txBase.Data
		}

		previewBase := &TxPreviewSimple{
			txBase.TxScript,
			txBase.DataVersion,
			dataPublic,
			nil,
			nil,
		}

		if txBase.HasVin() {
			previewBase.Vin = txBase.Vin.PublicKey
		}

		switch txBase.TxScript {
		case transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		case transaction_simple.SCRIPT_RESOLUTION_PAY_IN_FUTURE:

			txBaseExtra := txBase.Extra.(*transaction_simple_extra.TransactionSimpleExtraResolutionPayInFuture)

			previewBase.Extra = &TxPreviewSimpleExtraResolutionPayInFuture{
				txBaseExtra.TxId,
				txBaseExtra.PayloadIndex,
				txBaseExtra.Resolution,
			}
		}

		base = previewBase

	case transaction_type.TX_ZETHER:
		txBase := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
		payloads := make([]*TxPreviewZetherPayload, len(txBase.Payloads))
		for i, payload := range txBase.Payloads {

			var dataPublic []byte
			if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
				dataPublic = payload.Data
			}

			power, err := crypto.GetPowerof2(len(payload.Statement.C))
			if err != nil {
				return nil, err
			}

			var payloadExtra interface{}
			switch payload.PayloadScript {
			case transaction_zether_payload_script.SCRIPT_STAKING_REWARD:
				txPayloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraStakingReward)
				payloadExtra = &TxPreviewZetherPayloadExtraStakingReward{txPayloadExtra.Reward}
			case transaction_zether_payload_script.SCRIPT_STAKING:
				payloadExtra = &TxPreviewZetherPayloadExtraStaking{}
			case transaction_zether_payload_script.SCRIPT_SPEND:
				payloadExtra = &TxPreviewZetherPayloadExtraSpend{}
			case transaction_zether_payload_script.SCRIPT_PAY_IN_FUTURE:
				txPayloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraPayInFuture)
				payloadExtra = &TxPreviewZetherPayloadExtraPayToScript{txPayloadExtra.Deadline, txPayloadExtra.DefaultResolution, txPayloadExtra.MultisigThreshold}
			}

			payloads[i] = &TxPreviewZetherPayload{
				payload.PayloadScript,
				payload.Asset,
				payload.BurnValue,
				payload.DataVersion,
				dataPublic,
				byte(power),
				payloadExtra,
			}

		}

		previewBase := &TxPreviewZether{
			Payloads: payloads,
		}

		base = previewBase
	default:
		return nil, errors.New("Invalid tx.Version")
	}

	fee, err := tx.ComputeFee()
	if err != nil {
		return nil, err
	}

	return &TxPreview{
		base,
		tx.Version,
		tx.Bloom.Hash,
		fee,
	}, nil
}
