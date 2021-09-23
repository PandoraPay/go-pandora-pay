package config_fees

var (
	FEES_PER_BYTE        = uint64(10)
	FEES_PER_BYTE_ZETHER = uint64(20)
)

func ComputeTxSimpleFees(size, feePerByte uint64) uint64 {
	return size * feePerByte
}

func ComputeTxZetherFees(ringMembers, registrations, feePerByteZether uint64) uint64 {
	return ringMembers*feePerByteZether + registrations*feePerByteZether
}
