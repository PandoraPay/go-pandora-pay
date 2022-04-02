package addresses

type AddressVersion uint64

const (
	SIMPLE_PUBLIC_KEY_HASH            AddressVersion = 128
	SIMPLE_PUBLIC_KEY_HASH_INTEGRATED AddressVersion = 129
)

func (e AddressVersion) String() string {
	switch e {
	case SIMPLE_PUBLIC_KEY_HASH:
		return "SIMPLE_PUBLIC_KEY"
	case SIMPLE_PUBLIC_KEY_HASH_INTEGRATED:
		return "SIMPLE_PUBLIC_KEY_HASH_INTEGRATED"
	default:
		return "Unknown Address Version"
	}
}
