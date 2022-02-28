package txs_validator

import (
	"pandora-pay/blockchain/transactions/transaction"
	"time"
)

type txValidatedWork struct {
	wait       chan struct{}
	status     int32 //use atomic
	tx         *transaction.Transaction
	time       int64
	result     error
	bloomExtra any
}

const (
	EXPIRE_TIME_MS = 10 * 60 * time.Second
)

const (
	TX_VALIDATED_INIT int32 = iota
	TX_VALIDATED_PROCCESSED
)
