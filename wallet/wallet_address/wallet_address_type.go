package wallet_address

type Version int

const (
	VERSION_NORMAL Version = iota
)

func (e Version) String() string {
	switch e {
	case VERSION_NORMAL:
		return "VERSION_NORMAL"
	default:
		return "Unknown Wallet Address Version"
	}
}
