package wallet

import (
	"errors"
	"math/big"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type DecryptZetherPayloadOutput struct {
	WhisperSenderValid    bool             `json:"whisperSenderValid" msgpack:"whisperSenderValid"`
	SentAmount            uint64           `json:"sentAmount" msgpack:"sentAmount"`
	WhisperRecipientValid bool             `json:"whisperRecipientValid" msgpack:"whisperRecipientValid"`
	ReceivedAmount        uint64           `json:"receivedAmount" msgpack:"receivedAmount"`
	RecipientIndex        int              `json:"recipientIndex" msgpack:"recipientIndex"`
	Message               helpers.HexBytes `json:"message" msgpack:"message"`
}

type DecryptTxZether struct {
	Payloads []*DecryptZetherPayloadOutput `json:"payloads" msgpack:"payloads"`
}

type DecryptedTx struct {
	Type     transaction_type.TransactionVersion `json:"type" msgpack:"type"`
	ZetherTx *DecryptTxZether                    `json:"zetherTx" msgpack:"zetherTx"`
}

func (w *Wallet) DecryptTx(tx *transaction.Transaction) (*DecryptedTx, error) {

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
	case transaction_type.TX_ZETHER:
		txBase := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

		var data []byte
		output.ZetherTx = &DecryptTxZether{
			make([]*DecryptZetherPayloadOutput, len(txBase.Payloads)),
		}

		for t, payload := range txBase.Payloads {
			for i, publicKey := range txBase.Bloom.PublicKeyLists[t] {
				if addr := w.GetWalletAddressByPublicKey(publicKey, true); addr != nil {

					decyptedZetherPayload := &DecryptZetherPayloadOutput{
						RecipientIndex: -1,
					}
					output.ZetherTx.Payloads[t] = decyptedZetherPayload

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
									decyptedZetherPayload.WhisperSenderValid = true
									decyptedZetherPayload.SentAmount = amount
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
									decyptedZetherPayload.WhisperRecipientValid = true
									decyptedZetherPayload.ReceivedAmount = v1Value.Uint64()
								}
							}
						}
					}

					if output.ZetherTx.Payloads[t].WhisperSenderValid {
						rinputs := append([]byte{}, txBase.ChainHash...)
						for _, publicKey2 := range txBase.Bloom.PublicKeyLists[t] {
							rinputs = append(rinputs, publicKey2...)
						}

						rencrypted := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(append([]byte(crypto.PROTOCOL_CONSTANT), rinputs...))), secretPoint.BigInt())
						r := crypto.ReducedHash(rencrypted.EncodeCompressed())

						parity := payload.Proof.Parity()
						for k := range payload.Statement.C {
							if (k%2 == 0) == parity {
								continue
							}

							if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
								shared_key, err := crypto.GenerateSharedSecret(r, payload.Statement.Publickeylist[k])
								if err != nil {
									continue
								}

								data = append([]byte{}, payload.Data...)
								if err = crypto.EncryptDecryptUserData(cryptography.SHA3(append(shared_key, payload.Statement.Publickeylist[k].EncodeCompressed()...)), data); err != nil {
									continue
								}

								decyptedZetherPayload.Message = data
								decyptedZetherPayload.RecipientIndex = k
								break
							}

							var x bn256.G1
							x.ScalarMult(crypto.G, new(big.Int).SetUint64(decyptedZetherPayload.SentAmount-payload.Statement.Fee-payload.BurnValue))
							x.Add(new(bn256.G1).Set(&x), new(bn256.G1).ScalarMult(payload.Statement.Publickeylist[k], r))

							if x.String() == payload.Statement.C[k].String() {

								shared_key, err := crypto.GenerateSharedSecret(r, payload.Statement.Publickeylist[k])
								if err != nil {
									continue
								}

								if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
									data = append([]byte{}, payload.Data...)
									if err = crypto.EncryptDecryptUserData(cryptography.SHA3(append(shared_key, payload.Statement.Publickeylist[k].EncodeCompressed()...)), data); err != nil {
										continue
									}

									decyptedZetherPayload.Message = data
								} else if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
									decyptedZetherPayload.Message = payload.Data
								}

								output.ZetherTx.Payloads[t].RecipientIndex = k
							}

						}

					} else if decyptedZetherPayload.WhisperRecipientValid {

						shared_key, err := crypto.GenerateSharedSecret(secretPoint.BigInt(), payload.Statement.D)
						if err != nil {
							continue
						}

						if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
							data = append([]byte{}, payload.Data...)
							if err = crypto.EncryptDecryptUserData(cryptography.SHA3(append(shared_key, addr.PublicKey...)), data); err != nil {
								continue
							}
							decyptedZetherPayload.Message = data
						} else if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
							decyptedZetherPayload.Message = payload.Data
						}
					}

				}
			}
		}
	}

	return output, nil
}
