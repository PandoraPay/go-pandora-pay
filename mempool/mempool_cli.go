package mempool

import (
	"context"
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/gui"
	"time"
)

func (mempool *Mempool) initCLI() {

	cliShowTxs := func(cmd string, ctx context.Context) (err error) {

		transactions := mempool.Txs.GetTxsFromMap()
		if len(transactions) == 0 {
			return
		}

		gui.GUI.OutputWrite("Mempool Transactions:")
		for _, out := range transactions {
			switch out.Tx.Version {
			case transaction_type.TX_SIMPLE:
				txBase := out.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
				nonce := txBase.Nonce
				gui.GUI.OutputWrite(fmt.Sprintf("%12s %2d %7d %6d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC822), txBase.TxScript, nonce, out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
			case transaction_type.TX_ZETHER:
				txBase := out.Tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
				gui.GUI.OutputWrite(fmt.Sprintf("%12s %2d %7d %6d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC822), len(txBase.Payloads), txBase.ChainHeight, out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
			}
		}

		return
	}

	gui.GUI.CommandDefineCallback("Show Txs", cliShowTxs, true)
}
