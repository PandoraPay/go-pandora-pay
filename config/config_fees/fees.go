package config_fees

var (
	FEES_PER_BYTE = uint64(10)
)

func ComputeTxFees(size uint64, feePerByte uint64) uint64 {
	return size * feePerByte
}
