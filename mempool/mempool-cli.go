package mempool

import (
	"encoding/hex"
	"fmt"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/gui"
	"time"
)

func (mempool *Mempool) initCLI() {

	cliShowTxs := func(cmd string) (err error) {

		transactions := mempool.Txs.GetTxsList()
		if len(transactions) == 0 {
			return
		}

		gui.GUI.OutputWrite("Mempool Transactions:")
		for _, out := range transactions {
			if out.Tx.TxType == transaction_type.TX_SIMPLE {
				txBase := out.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
				nonce := txBase.Nonce
				gui.GUI.OutputWrite(fmt.Sprintf("%12s %4d %7d %6d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC822), txBase.TxScript, nonce, out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
			}
		}

		return
	}

	gui.GUI.CommandDefineCallback("Show Txs", cliShowTxs, true)
}
