package transactions_builder

import (
	"pandora-pay/blockchain/data/tokens/token"
	"pandora-pay/transactions-builder/wizard"
)

type TransactionsBuilderFeeFloat struct {
	Fixed       float64 `json:"fixed,omitempty"`
	PerByte     float64 `json:"perByte,omitempty"`
	PerByteAuto bool    `json:"perByteAuto,omitempty"`
}

type TransactionSimpleOutputFloat struct {
	Amount                float64 `json:"amount,omitempty"`
	PublicKey             []byte  `json:"publicKey,omitempty"`
	HasRegistration       bool    `json:"hasRegistration,omitempty"`
	RegistrationSignature []byte  `json:"registrationSignature,omitempty"`
}

func (fee *TransactionsBuilderFeeFloat) convertToWizardFee(tok *token.Token) (*wizard.TransactionsWizardFee, error) {

	var err error

	out := &wizard.TransactionsWizardFee{
		PerByteAuto: fee.PerByteAuto,
	}

	if fee.Fixed > 0 {
		if out.Fixed, err = tok.ConvertToUnits(fee.Fixed); err != nil {
			return nil, err
		}
	}
	if fee.PerByte > 0 {
		if out.PerByte, err = tok.ConvertToUnits(fee.PerByte); err != nil {
			return nil, err
		}
	}

	return out, nil
}
