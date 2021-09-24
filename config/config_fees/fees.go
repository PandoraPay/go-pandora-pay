package config_fees

var (
	FEES_PER_BYTE        = uint64(10)
	FEES_PER_BYTE_ZETHER = uint64(20)
)

func ComputeTxFees(size, feePerByte uint64) uint64 {
	return size * feePerByte
}
