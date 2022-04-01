package addresses

type AddressVersion uint64

const (
	SIMPLE_PUBLIC_KEY_HASH AddressVersion = iota
)

func (e AddressVersion) String() string {
	switch e {
	case SIMPLE_PUBLIC_KEY_HASH:
		return "SIMPLE_PUBLIC_KEY"
	default:
		return "Unknown Address Version"
	}
}
