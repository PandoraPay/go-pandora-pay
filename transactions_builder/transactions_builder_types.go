package transactions_builder

import (
	"pandora-pay/blockchain/data_storage/tokens/token"
	"pandora-pay/transactions_builder/wizard"
)

type TransactionsBuilderFeeFloat struct {
	Fixed       float64 `json:"fixed,omitempty"`
	PerByte     float64 `json:"perByte,omitempty"`
	PerByteAuto bool    `json:"perByteAuto,omitempty"`
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
