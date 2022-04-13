package wallet

type Version int

const (
	VERSION_SIMPLE Version = iota
	VERSION_SIMPLE_HARDENED
)

func (e Version) String() string {
	switch e {
	case VERSION_SIMPLE:
		return "VERSION_SIMPLE"
	case VERSION_SIMPLE_HARDENED:
		return "VERSION_SIMPLE_HARDENED"
	default:
		return "Unknown Version"
	}
}

type EncryptedVersion int

const (
	ENCRYPTED_VERSION_PLAIN_TEXT EncryptedVersion = iota
	ENCRYPTED_VERSION_ENCRYPTION_ARGON2
)

func (e EncryptedVersion) String() string {
	switch e {
	case ENCRYPTED_VERSION_PLAIN_TEXT:
		return "ENCRYPTED_VERSION_PLAIN_TEXT"
	case ENCRYPTED_VERSION_ENCRYPTION_ARGON2:
		return "ENCRYPTED_VERSION_ENCRYPTION_ARGON2"
	default:
		return "Unknown EncryptedVersion"
	}
}
