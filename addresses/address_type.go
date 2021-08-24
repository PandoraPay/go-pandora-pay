package addresses

type AddressVersion uint64

const (
	SIMPLE_PUBLIC_KEY AddressVersion = 1
)

func (e AddressVersion) String() string {
	switch e {
	case SIMPLE_PUBLIC_KEY:
		return "SIMPLE_PUBLIC_KEY"
	default:
		return "Unknown Address Version"
	}
}
