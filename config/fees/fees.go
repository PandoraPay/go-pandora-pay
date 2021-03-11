package fees

import "pandora-pay/config"

var (
	FEES_PER_BYTE = map[[20]byte]uint64{
		config.NATIVE_TOKEN_FULL: 10,
	}
)

func ComputeTxFees(size uint64, feePerByte uint64) uint64 {
	return size * feePerByte
}
