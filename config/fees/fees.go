package fees

import "pandora-pay/config"

var (
	FEES_PER_BYTE = map[string]uint64{
		string(config.NATIVE_TOKEN): 10,
	}
)

func ComputeTxFees(size uint64, feePerByte uint64) uint64 {
	return size * feePerByte
}
