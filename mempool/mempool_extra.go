package mempool

import (
	"bytes"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"sort"
)

type ContinueProcessingType byte

const (
	CONTINUE_PROCESSING_ERROR ContinueProcessingType = iota
	CONTINUE_PROCESSING_NO_ERROR_RESET
	CONTINUE_PROCESSING_NO_ERROR
)

func (mempool *Mempool) ExistsTxSimpleVersion(publicKey []byte, version transaction_simple.ScriptType) bool {

	txs := mempool.Txs.GetTxsFromMap()
	if txs == nil {
		return false
	}

	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if base.TxScript == version && bytes.Equal(base.Vin.PublicKey, publicKey) {
				return true
			}
		}
	}
	return false
}

func (mempool *Mempool) ExistsTxZetherVersion(publicKey []byte, version transaction_zether_payload.PayloadScriptType) bool {

	txs := mempool.Txs.GetTxsFromMap()
	if txs == nil {
		return false
	}

	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_ZETHER {
			base := tx.Tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
			for _, payload := range base.Payloads {
				if payload.PayloadScript == version {
					for _, publicKeyPoint := range payload.Statement.Publickeylist {
						if bytes.Equal(publicKeyPoint.EncodeCompressed(), publicKey) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (mempool *Mempool) CountInputTxs(publicKey []byte) uint64 {

	txs := mempool.Txs.GetTxsFromMap()

	count := uint64(0)
	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin.PublicKey, publicKey) {
				count++
			}
		}
		if tx.Tx.Version == transaction_type.TX_ZETHER {
			base := tx.Tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
			for _, payload := range base.Payloads {

				for _, payloadPoint := range payload.Statement.Publickeylist {
					txPublicKey := payloadPoint.EncodeCompressed()
					if bytes.Equal(publicKey, txPublicKey) {
						count++
					}
				}

			}
		}
	}

	return count
}

func (mempool *Mempool) GetNonce(publicKey []byte, nonce uint64) uint64 {

	txs := mempool.Txs.GetTxsFromMap()

	nonces := make(map[uint64]bool)
	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin.PublicKey, publicKey) {
				nonces[base.Nonce] = true
			}
		}
	}

	for {
		if nonces[nonce] {
			nonce += 1
		} else {
			break
		}
	}

	return nonce
}

func (mempool *Mempool) GetZetherBalance(publicKey []byte, balanceInit []byte) ([]byte, error) {
	result, err := mempool.GetZetherBalanceMultiple([][]byte{publicKey}, [][]byte{balanceInit})
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

func (mempool *Mempool) GetZetherBalanceMultiple(publicKeys [][]byte, balancesInit [][]byte) ([][]byte, error) {

	txs := mempool.Txs.GetTxsFromMap()
	var balance *crypto.ElGamal
	var err error
	var acckey crypto.Point

	output := make([][]byte, len(publicKeys))
	for i, publicKey := range publicKeys {

		balanceInit := balancesInit[i]

		if balanceInit == nil {
			if err = acckey.DecodeCompressed(publicKey); err != nil {
				return nil, err
			}
			balance = crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
		} else {
			if balance, err = new(crypto.ElGamal).Deserialize(helpers.CloneBytes(balanceInit)); err != nil {
				return nil, err
			}
		}

		changed := false
		for _, tx := range txs {
			if tx.Tx.Version == transaction_type.TX_ZETHER {
				base := tx.Tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
				for _, payload := range base.Payloads {
					for i, publicKeyPoint := range payload.Statement.Publickeylist {
						txPublicKey := publicKeyPoint.EncodeCompressed()
						if bytes.Equal(publicKey, txPublicKey) {
							echanges := crypto.ConstructElGamal(payload.Statement.C[i], payload.Statement.D)
							balance = balance.Add(echanges) // homomorphic addition of changes
							changed = true
						}
					}
				}
			}
		}

		if changed || balanceInit != nil {
			output[i] = balance.Serialize()
		}
	}

	return output, nil
}

func (mempool *Mempool) GetNextTransactionsToInclude(chainHash []byte) (out []*transaction.Transaction, outChainHash []byte) {

	result := mempool.result.Load()
	if result != nil {

		res := result.(*MempoolResult)

		if chainHash == nil || bytes.Equal(res.chainHash, chainHash) {
			txs := res.txs.Load().([]*mempoolTx)
			finalTxs := make([]*transaction.Transaction, len(txs))
			for i, tx := range txs {
				finalTxs[i] = tx.Tx
			}
			return finalTxs, res.chainHash
		}
	}

	return []*transaction.Transaction{}, nil
}

func sortTxs(txList []*mempoolTx) {
	sort.Slice(txList, func(i, j int) bool {

		if txList[i].FeePerByte == txList[j].FeePerByte && txList[i].Tx.Version == transaction_type.TX_SIMPLE && txList[j].Tx.Version == transaction_type.TX_SIMPLE {
			return txList[i].Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce < txList[j].Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce
		}

		return txList[i].FeePerByte < txList[j].FeePerByte
	})
}
