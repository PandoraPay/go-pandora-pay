package wallet

type Version int

const (
	VERSION_SIMPLE Version = 0
)

func (e Version) String() string {
	switch e {
	case VERSION_SIMPLE:
		return "VERSION_SIMPLE"
	default:
		return "Unknown Version"
	}
}

type EncryptedVersion int

const (
	ENCRYPTED_VERSION_PLAIN_TEXT EncryptedVersion = iota
	ENCRYPTED_VERSION_ENCRYPTION
)

func (e EncryptedVersion) String() string {
	switch e {
	case ENCRYPTED_VERSION_PLAIN_TEXT:
		return "ENCRYPTED_VERSION_PLAIN_TEXT"
	case ENCRYPTED_VERSION_ENCRYPTION:
		return "ENCRYPTED_VERSION_ENCRYPTION"
	default:
		return "Unknown EncryptedVersion"
	}
}
