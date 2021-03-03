package fees

const (
	FEE_PER_BYTE_DEFAULT = 1000
)

func ComputeTxFees(size uint64, feePerByte uint64) uint64 {
	return size * feePerByte
}
