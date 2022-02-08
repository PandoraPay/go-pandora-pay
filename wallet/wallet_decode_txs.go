package wallet

import (
	"context"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
)

func (w *Wallet) DecodeZetherTx(tx *transaction.Transaction, ctx context.Context) error {

	if tx == nil {
		return errors.New("Transaction is invalid")
	}
	if tx.Version != transaction_type.TX_ZETHER {
		return errors.New("Transaction is not zether")
	}

	if err := tx.BloomAll(); err != nil {
		return err
	}

	txBase := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

	for t, payload := range txBase.Payloads {
		for i, publicKey := range txBase.Bloom.PublicKeyLists[t] {
			if addr := w.GetWalletAddressByPublicKey(publicKey, true); addr != nil {

				echanges := crypto.ConstructElGamal(payload.Statement.C[i], payload.Statement.D)
				balance := crypto.ConstructElGamal(payload.Statement.CLn[i], payload.Statement.CRn[i])
				echanges = echanges.Neg()
				initBalance := balance.Add(balance)

				previousBalance, err := w.DecodeBalanceByPublicKey(publicKey, initBalance, payload.Asset, false, 0, false, true, ctx, nil)
				if err != nil {
					return err
				}

				gui.GUI.Log("-------------------------------")
				gui.GUI.Log("previousBalance", previousBalance)
				gui.GUI.Log("-------------------------------")

				currentBalance, err := w.DecodeBalanceByPublicKey(publicKey, initBalance.Add(echanges), payload.Asset, true, previousBalance, false, true, ctx, nil)
				if err != nil {
					return err
				}

				gui.GUI.Log("-------------------------------")
				gui.GUI.Log("currentBalance", currentBalance)
				gui.GUI.Log("-------------------------------")
			}
		}
	}

	return nil
}
