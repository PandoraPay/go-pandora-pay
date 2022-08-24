package builds_data

import (
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
	"pandora-pay/txs_builder/wizard"
)

type WalletInitializeBalanceDecryptorReq struct {
	TableSize int `json:"tableSize"`
}

type WalletDecryptBalanceReq struct {
	PublicKey     []byte `json:"publicKey"`
	PrivateKey    []byte `json:"privateKey"`
	PreviousValue uint64 `json:"previousValue"`
	Balance       []byte `json:"balance"`
	Asset         []byte `json:"asset"`
}

type zetherTxDataSender struct {
	PrivateKey       []byte `json:"privateKey"`
	SpendPrivateKey  []byte `json:"spendPrivateKey"`
	DecryptedBalance uint64 `json:"decryptedBalance"`
}

type zetherTxDataPayloadBase struct {
	Sender               *zetherTxDataSender                                 `json:"sender"`
	Asset                []byte                                              `json:"asset"`
	Amount               uint64                                              `json:"amount"`
	Recipient            string                                              `json:"recipient"`
	Burn                 uint64                                              `json:"burn"`
	RingSenderMembers    []string                                            `json:"ringSenderMembers"`
	RingRecipientMembers []string                                            `json:"ringRecipientMembers"`
	Data                 *wizard.WizardTransactionData                       `json:"data"`
	Fees                 *wizard.WizardZetherTransactionFee                  `json:"fees"`
	ScriptType           transaction_zether_payload_script.PayloadScriptType `json:"scriptType"`
	Extra                wizard.WizardZetherPayloadExtra                     `json:"extra"`
}

type TransactionsBuilderCreateZetherTxReq struct {
	Payloads          []*zetherTxDataPayloadBase   `json:"payloads"`
	Accs              map[string]map[string][]byte `json:"accs"`
	Regs              map[string][]byte            `json:"regs"`
	ChainKernelHeight uint64                       `json:"chainKernelHeight"`
	ChainKernelHash   []byte                       `json:"chainKernelHash"`
}
