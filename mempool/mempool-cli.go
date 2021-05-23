package mempool

import (
	"encoding/hex"
	"fmt"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/context"
	"time"
)

func (mempool *Mempool) initCLI() {

	cliShowTxs := func(cmd string) (err error) {

		transactions := mempool.GetTxsList()
		if len(transactions) == 0 {
			return
		}

		context.GUI.OutputWrite("Mempool Transactions:")
		for _, out := range transactions {
			if out.Tx.TxType == transaction_type.TxSimple {
				txBase := out.Tx.TxBase.(*transaction_simple.TransactionSimple)
				nonce := txBase.Nonce
				context.GUI.OutputWrite(fmt.Sprintf("%12s %4d %7d %6d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC822), txBase.TxScript, nonce, out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
			}
		}

		return
	}

	context.GUI.CommandDefineCallback("Show Txs", cliShowTxs)
}
