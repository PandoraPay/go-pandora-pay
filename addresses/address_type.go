package addresses

type AddressVersion uint64

const (
	SIMPLE_PUBLIC_KEY AddressVersion = iota
	SIMPLE_DELEGATED
)

func (e AddressVersion) String() string {
	switch e {
	case SIMPLE_PUBLIC_KEY:
		return "SIMPLE_PUBLIC_KEY"
	case SIMPLE_DELEGATED:
		return "SIMPLE_DELEGATED"
	default:
		return "Unknown Address Version"
	}
}
