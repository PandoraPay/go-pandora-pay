package wallet

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
)

type DecryptZetherPayloadOutput struct {
	WhisperSenderValid    bool   `json:"whisperSenderValid" msgpack:"whisperSenderValid"`
	SentAmount            uint64 `json:"sentAmount" msgpack:"sentAmount"`
	WhisperRecipientValid bool   `json:"whisperRecipientValid" msgpack:"whisperRecipientValid"`
	Blinder               []byte `json:"blinder" msgpack:"blinder"`
	ReceivedAmount        uint64 `json:"receivedAmount" msgpack:"receivedAmount"`
	RecipientIndex        int    `json:"recipientIndex" msgpack:"recipientIndex"`
	Message               []byte `json:"message" msgpack:"message"`
}

type DecryptTxZether struct {
	Payloads []*DecryptZetherPayloadOutput `json:"payloads" msgpack:"payloads"`
}

type DecryptedTx struct {
	Type     transaction_type.TransactionVersion `json:"type" msgpack:"type"`
	ZetherTx *DecryptTxZether                    `json:"zetherTx" msgpack:"zetherTx"`
}

func (w *Wallet) DecryptTx(tx *transaction.Transaction, publicKeyHash []byte) (*DecryptedTx, error) {

	if tx == nil {
		return nil, errors.New("Transaction is invalid")
	}
	if err := tx.BloomAll(); err != nil {
		return nil, err
	}

	output := &DecryptedTx{
		Type: tx.Version,
	}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
	}

	return output, nil
}
