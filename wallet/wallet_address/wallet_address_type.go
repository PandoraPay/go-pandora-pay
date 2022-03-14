package wallet_address

type Version int

const (
	VERSION_NORMAL Version = iota
	VERSION_DELEGATED_STAKE
)

func (e Version) String() string {
	switch e {
	case VERSION_NORMAL:
		return "VERSION_NORMAL"
	case VERSION_DELEGATED_STAKE:
		return "VERSION_DELEGATED_STAKE"
	default:
		return "Unknown Wallet Address Version"
	}
}
