package wallet_address

type Version int

const (
	VERSION_TRANSPARENT Version = 0
)

func (e Version) String() string {
	switch e {
	case VERSION_TRANSPARENT:
		return "VERSION_TRANSPARENT"
	default:
		return "Unknown Wallet Address Version"
	}
}
