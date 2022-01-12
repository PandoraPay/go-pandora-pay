package wizard

import (
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type WizardTransactionFee struct {
	Fixed             uint64 `json:"fixed,omitempty" msgpack:"fixed,omitempty"`
	PerByte           uint64 `json:"perByte,omitempty" msgpack:"perByte,omitempty"`
	PerByteExtraSpace uint64 `json:"perByteExtraSpace,omitempty" msgpack:"perByteExtraSpace,omitempty"`
	PerByteAuto       bool   `json:"perByteAuto,omitempty" msgpack:"perByteAuto,omitempty"`
}

func (fee *WizardTransactionFee) Clone() *WizardTransactionFee {
	return &WizardTransactionFee{
		Fixed:             fee.Fixed,
		PerByte:           fee.PerByte,
		PerByteExtraSpace: fee.PerByteExtraSpace,
		PerByteAuto:       fee.PerByteAuto,
	}
}

type WizardTransactionData struct {
	Data    helpers.HexBytes `json:"data,omitempty" msgpack:"data,omitempty"`
	Encrypt bool             `json:"encrypt,omitempty" msgpack:"encrypt,omitempty"`
}

func (data *WizardTransactionData) getDataVersion() transaction_data.TransactionDataVersion {
	if data.Data == nil || len(data.Data) == 0 {
		return transaction_data.TX_DATA_NONE
	}
	if data.Encrypt {
		return transaction_data.TX_DATA_ENCRYPTED
	}
	return transaction_data.TX_DATA_PLAIN_TEXT
}

func (data *WizardTransactionData) getData() ([]byte, error) {
	if len(data.Data) == 0 {
		return nil, nil
	}
	if !data.Encrypt {
		return data.Data, nil
	} else {

		panic("not implemented")
		//pub, err := ecdsa.DecompressPubkey(data.PublicKeyToEncrypt)
		//if err != nil {
		//	return nil, err
		//}
		//
		//return ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(pub), data.Data, nil, nil)
	}
}
