package wallet

type Version int

const (
	VersionSimple Version = 0
)

func (e Version) String() string {
	switch e {
	case VersionSimple:
		return "VersionSimple"
	default:
		return "Unknown Version"
	}
}

type EncryptedVersion int

const (
	EncryptedVersionPlainText EncryptedVersion = iota
	EncryptedVersionEncryption
)

func (e EncryptedVersion) String() string {
	switch e {
	case EncryptedVersionPlainText:
		return "EncryptedVersionPlainText"
	case EncryptedVersionEncryption:
		return "EncryptedVersionEncryption"
	default:
		return "Unknown EncryptedVersion"
	}
}
