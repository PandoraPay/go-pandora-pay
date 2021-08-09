package transactions_builder

import (
	"pandora-pay/blockchain/tokens"
	"pandora-pay/transactions-builder/wizard"
)

type TransactionsBuilderFeeFloat struct {
	Fixed       float64 `json:"fixed,omitempty"`
	PerByte     float64 `json:"perByte,omitempty"`
	PerByteAuto bool    `json:"perByteAuto,omitempty"`
	Token       []byte  `json:"token,omitempty"`
}

type TransactionsBuilderFeeFloatExtra struct {
	TransactionsBuilderFeeFloat
	PayInExtra bool `json:"payInExtra,omitempty"`
}

func (fee *TransactionsBuilderFeeFloat) convertToWizardFee(toks *tokens.Tokens) (*wizard.TransactionsWizardFee, error) {

	tok, err := toks.GetTokenRequired(fee.Token)
	if err != nil {
		return nil, err
	}

	out := &wizard.TransactionsWizardFee{
		PerByteAuto: fee.PerByteAuto,
		Token:       fee.Token,
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

func (fee *TransactionsBuilderFeeFloatExtra) convertToWizardFee(toks *tokens.Tokens) (*wizard.TransactionsWizardFeeExtra, error) {
	out, err := fee.TransactionsBuilderFeeFloat.convertToWizardFee(toks)
	if err != nil {
		return nil, err
	}
	return &wizard.TransactionsWizardFeeExtra{
		*out,
		fee.PayInExtra,
	}, nil
}
