package address_balance_decryptor

type addressBalanceDecryptedWork struct {
	wait      chan struct{}
	status    int32 //use atomic
	time      int64
	decrypted uint64
	result    error
}

const (
	ADDRESS_BALANCE_DECRYPTED_INIT int32 = iota
	ADDRESS_BALANCE_DECRYPTED_PROCESSED
)
