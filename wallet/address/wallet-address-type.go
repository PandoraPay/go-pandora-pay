package wallet_address

type Version int

const (
	VersionTransparent Version = 0
)

func (e Version) String() string {
	switch e {
	case VersionTransparent:
		return "VersionTransparent"
	default:
		return "Unknown Wallet Address Version"
	}
}
