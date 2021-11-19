package addresses

type AddressVersion uint64

const (
	SIMPLE_PUBLIC_KEY AddressVersion = iota
	SIMPLE_PUBLIC_KEY_REGISTRATION
)

func (e AddressVersion) String() string {
	switch e {
	case SIMPLE_PUBLIC_KEY:
		return "SIMPLE_PUBLIC_KEY"
	default:
		return "Unknown Address Version"
	}
}
