package addresses

type PrivateKeyVersion uint64

const (
	SIMPLE_PRIVATE_KEY PrivateKeyVersion = iota
	SIMPLE_PRIVATE_KEY_WIF
)

func (e PrivateKeyVersion) String() string {
	switch e {
	case SIMPLE_PRIVATE_KEY:
		return "SIMPLE_PRIVATE_KEY"
	case SIMPLE_PRIVATE_KEY_WIF:
		return "SIMPLE_PRIVATE_KEY_WIF"
	default:
		return "Unknown Address Version"
	}
}
