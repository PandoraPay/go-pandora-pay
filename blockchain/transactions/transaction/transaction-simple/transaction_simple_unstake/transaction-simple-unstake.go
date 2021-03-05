package transaction_simple_unstake

import (
	"pandora-pay/helpers"
)

type TransactionSimpleUnstake struct {
	UnstakeAmount   uint64
	UnstakeFeeExtra uint64
}

//func (tx *TransactionSimpleUnstake) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) {
//	if !acc.HasDelegatedStake() {
//		panic("Account has no delegated stake")
//	}
//	acc.AddDelegatedStake( false, tx.UnstakeAmount, blockHeight );
//	if err = acc.AddDelegatedStake( false, tx.UnstakeFeeExtra, blockHeight ); err != nil {
//		return
//	}
//	return
//}
//
//func (tx *TransactionSimpleUnstake) RemoveTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) {
//	if err = acc.AddBalance( true, tx.UnstakeFeeExtra, config.NATIVE_TOKEN ); err != nil {
//		return
//	}
//	return
//}

func (tx *TransactionSimpleUnstake) Validate() {
	if tx.UnstakeAmount == 0 {
		panic("Unstake must be greather than zero")
	}
}

func (tx *TransactionSimpleUnstake) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.UnstakeAmount)
	writer.WriteUvarint(tx.UnstakeFeeExtra)
}

func (tx *TransactionSimpleUnstake) Deserialize(reader *helpers.BufferReader) {
	tx.UnstakeAmount = reader.ReadUvarint()
	tx.UnstakeFeeExtra = reader.ReadUvarint()
}
