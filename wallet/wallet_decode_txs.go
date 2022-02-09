package wallet

import (
	"errors"
	"math/big"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type PayloadOutput struct {
	WhisperSenderValid    bool   `json:"whisperSenderValid" msgpack:"whisperSenderValid"`
	SentAmount            uint64 `json:"sentAmount" msgpack:"sentAmount"`
	WhisperRecipientValid bool   `json:"whisperRecipientValid" msgpack:"whisperRecipientValid"`
	ReceivedAmount        uint64 `json:"receivedAmount" msgpack:"receivedAmount"`
}

func (w *Wallet) DecodeZetherTx(tx *transaction.Transaction) ([]*PayloadOutput, error) {

	if tx == nil {
		return nil, errors.New("Transaction is invalid")
	}
	if tx.Version != transaction_type.TX_ZETHER {
		return nil, errors.New("Transaction is not zether")
	}

	if err := tx.BloomAll(); err != nil {
		return nil, err
	}

	txBase := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

	output := make([]*PayloadOutput, len(txBase.Payloads))

	for t, payload := range txBase.Payloads {
		for i, publicKey := range txBase.Bloom.PublicKeyLists[t] {
			if addr := w.GetWalletAddressByPublicKey(publicKey, true); addr != nil {

				output[t] = &PayloadOutput{}

				echanges := crypto.ConstructElGamal(payload.Statement.C[i], payload.Statement.D)
				secretPoint := new(crypto.BNRed).SetBytes(addr.PrivateKey.Key)

				//check sender whisper
				v2Computed := crypto.ReducedHash(new(bn256.G1).ScalarMult(payload.Statement.D, secretPoint.BigInt()).EncodeCompressed())
				v2Sub := new(big.Int).Sub(new(big.Int).SetBytes(payload.WhisperSender), v2Computed)
				v2Value := new(big.Int).Mod(v2Sub, bn256.Order)

				if v2Value.IsUint64() {
					amount := v2Value.Uint64()
					if err := helpers.SafeUint64Add(&amount, payload.Statement.Fee); err == nil {
						if err := helpers.SafeUint64Add(&amount, payload.BurnValue); err == nil {
							if addr.PrivateKey.CheckMatchBalanceDecoded(echanges.Neg(), amount) {
								output[t].WhisperSenderValid = true
								output[t].SentAmount = v2Value.Uint64()
							}
						}
					}
				}

				//check recipient whisper
				v1Computed := crypto.ReducedHash(new(bn256.G1).ScalarMult(payload.Statement.D, secretPoint.BigInt()).EncodeCompressed())
				v1Sub := new(big.Int).Sub(new(big.Int).SetBytes(payload.WhisperRecipient), v1Computed)
				v1Value := new(big.Int).Mod(v1Sub, bn256.Order)

				if v1Value.IsUint64() {
					amount := v1Value.Uint64()
					if err := helpers.SafeUint64Add(&amount, payload.Statement.Fee); err == nil {
						if err := helpers.SafeUint64Add(&amount, payload.BurnValue); err == nil {
							if addr.PrivateKey.CheckMatchBalanceDecoded(echanges, amount) {
								output[t].WhisperRecipientValid = true
								output[t].ReceivedAmount = v1Value.Uint64()
							}
						}
					}

				}

			}
		}
	}

	return output, nil
}
