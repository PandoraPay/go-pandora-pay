package config_fees

var (
	FEES_PER_BYTE             = uint64(10)
	FEES_PER_BYTE_ZETHER      = uint64(20)
	FEES_PER_BYTE_EXTRA_SPACE = uint64(100)
)

func ComputeTxFees(size, feePerByte, extraSpace, feePerByeExtraSpace uint64) uint64 {
	return size*feePerByte + extraSpace*feePerByeExtraSpace
}
