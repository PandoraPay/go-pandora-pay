package info

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TxPreviewSimpleExtraUnstake struct {
	Amount uint64 `json:"amount" msgpack:"amount"`
}

type TxPreviewSimpleExtraUpdateDelegate struct {
	DelegatedStakingClaimAmount uint64 `json:"delegatedStakingClaimAmount" msgpack:"delegatedStakingClaimAmount"`
}

type TxPreviewSimple struct {
	Extra       interface{}                             `json:"extra" msgpack:"extra"`
	TxScript    transaction_simple.ScriptType           `json:"txScript" msgpack:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion" msgpack:"dataVersion"`
	DataPublic  helpers.HexBytes                        `json:"dataPublic" msgpack:"dataPublic"`
	Vin         helpers.HexBytes                        `json:"vin" msgpack:"vin"`
}

type TxPreviewZetherPayloadExtraClaim struct {
	DelegatePublicKey           helpers.HexBytes `json:"delegatePublicKey" msgpack:"delegatePublicKey"`
	DelegatedStakingClaimAmount uint64           `json:"delegatedStakingClaimAmount" msgpack:"delegatedStakingClaimAmount"`
}

type TxPreviewZetherPayloadExtraDelegateStake struct {
	DelegatePublicKey helpers.HexBytes `json:"delegatePublicKey" msgpack:"delegatePublicKey"`
}

type TxPreviewZetherPayload struct {
	PayloadScript transaction_zether_payload.PayloadScriptType `json:"payloadScript" msgpack:"payloadScript"`
	Asset         helpers.HexBytes                             `json:"asset" msgpack:"asset"`
	BurnValue     uint64                                       `json:"burnValue" msgpack:"burnValue"`
	DataVersion   transaction_data.TransactionDataVersion      `json:"dataVersion" msgpack:"dataVersion"`
	DataPublic    helpers.HexBytes                             `json:"dataPublic" msgpack:"dataPublic"`
	Ring          byte                                         `json:"ring" msgpack:"ring"`
	Extra         interface{}                                  `json:"extra" msgpack:"extra"`
}

type TxPreviewZether struct {
	Payloads []*TxPreviewZetherPayload `json:"payloads"  msgpack:"payloads"`
}

type TxPreview struct {
	TxBase  interface{}                         `json:"base"  msgpack:"base"`
	Version transaction_type.TransactionVersion `json:"version"  msgpack:"version"`
	Hash    helpers.HexBytes                    `json:"hash"  msgpack:"hash"`
	Fee     uint64                              `json:"fee"  msgpack:"fee"`
}

func CreateTxPreviewFromTx(tx *transaction.Transaction) (*TxPreview, error) {

	var base interface{}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		var baseExtra interface{}
		switch txBase.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE: //nothing to be copied
			txExtra := txBase.Extra.(*transaction_simple_extra.TransactionSimpleExtraUpdateDelegate)
			baseExtra = &TxPreviewSimpleExtraUpdateDelegate{txExtra.DelegatedStakingClaimAmount}
		case transaction_simple.SCRIPT_UNSTAKE:
			txExtra := txBase.Extra.(*transaction_simple_extra.TransactionSimpleExtraUnstake)
			baseExtra = &TxPreviewSimpleExtraUnstake{txExtra.Amount}
		}

		var dataPublic helpers.HexBytes
		if txBase.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
			dataPublic = txBase.Data
		}

		base = &TxPreviewSimple{
			baseExtra,
			txBase.TxScript,
			txBase.DataVersion,
			dataPublic,
			txBase.Vin.PublicKey,
		}

	case transaction_type.TX_ZETHER:
		txBase := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
		payloads := make([]*TxPreviewZetherPayload, len(txBase.Payloads))
		for i, payload := range txBase.Payloads {

			var dataPublic helpers.HexBytes
			if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
				dataPublic = payload.Data
			}

			power, err := crypto.GetPowerof2(len(payload.Statement.C))
			if err != nil {
				return nil, err
			}

			var payloadExtra interface{}
			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_CLAIM:
				txPayloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraClaim)
				payloadExtra = &TxPreviewZetherPayloadExtraClaim{txPayloadExtra.DelegatePublicKey, txPayloadExtra.DelegatedStakingClaimAmount}
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				txPayloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake)
				payloadExtra = &TxPreviewZetherPayloadExtraDelegateStake{txPayloadExtra.DelegatePublicKey}
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

		base = &TxPreviewZether{
			Payloads: payloads,
		}
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
