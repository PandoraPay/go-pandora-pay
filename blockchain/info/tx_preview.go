package info

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
)

type TxPreviewSimpleVin struct {
	PublicKey []byte `json:"publicKey" msgpack:"publicKey"`
	Amount    uint64 `json:"amount" msgpack:"amount"`
	Asset     []byte `json:"asset" msgpack:"asset"`
}

type TxPreviewSimpleVout struct {
	PublicKeyHash []byte `json:"publicKeyHash" msgpack:"publicKeyHash"`
	Amount        uint64 `json:"amount" msgpack:"amount"`
	Asset         []byte `json:"asset" msgpack:"asset"`
}

type TxPreviewSimple struct {
	Extra       interface{}                             `json:"extra" msgpack:"extra"`
	TxScript    transaction_simple.ScriptType           `json:"txScript" msgpack:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion" msgpack:"dataVersion"`
	DataPublic  []byte                                  `json:"dataPublic" msgpack:"dataPublic"`
	Vin         []*TxPreviewSimpleVin                   `json:"vin" msgpack:"vin"`
	Vout        []*TxPreviewSimpleVout                  `json:"vout" msgpack:"vout"`
}

type TxPreview struct {
	TxBase  interface{}                         `json:"base"  msgpack:"base"`
	Version transaction_type.TransactionVersion `json:"version"  msgpack:"version"`
	Hash    []byte                              `json:"hash"  msgpack:"hash"`
	Fee     uint64                              `json:"fee"  msgpack:"fee"`
}

func CreateTxPreviewFromTx(tx *transaction.Transaction) (*TxPreview, error) {

	var base interface{}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		var baseExtra interface{}

		var dataPublic []byte
		if txBase.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
			dataPublic = txBase.Data
		}

		previewVin := make([]*TxPreviewSimpleVin, len(txBase.Vin))
		for i, vin := range txBase.Vin {
			previewVin[i] = &TxPreviewSimpleVin{
				vin.PublicKey,
				vin.Amount,
				vin.Asset,
			}
		}

		previewVout := make([]*TxPreviewSimpleVout, len(txBase.Vout))
		for i, vout := range txBase.Vout {
			previewVout[i] = &TxPreviewSimpleVout{
				vout.PublicKeyHash,
				vout.Amount,
				vout.Asset,
			}
		}

		base = &TxPreviewSimple{
			baseExtra,
			txBase.TxScript,
			txBase.DataVersion,
			dataPublic,
			previewVin,
			previewVout,
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
