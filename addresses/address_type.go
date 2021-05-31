package addresses

type AddressVersion uint64

const (
	SIMPLE_PUBLIC_KEY_HASH AddressVersion = 0
	SIMPLE_PUBLIC_KEY      AddressVersion = 1
)

func (e AddressVersion) String() string {
	switch e {
	case SIMPLE_PUBLIC_KEY_HASH:
		return "SIMPLE_PUBLIC_KEY_HASH"
	case SIMPLE_PUBLIC_KEY:
		return "SIMPLE_PUBLIC_KEY"
	default:
		return "Unknown Address Version"
	}
}
