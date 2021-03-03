package fees

func ComputeTxFees(size, blockHeight uint64) uint64 {
	return size * 10
}
