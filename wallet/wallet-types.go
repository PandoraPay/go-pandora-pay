package wallet

type WalletVersion int

const (
	WalletVersionSimple WalletVersion = 0
)

func (e WalletVersion) String() string {
	switch e {
	case WalletVersionSimple:
		return "WalletVersionSimple"
	default:
		return "Unknown Version"
	}
}

type WalletEncryptedVersion int

const (
	WalletEncryptedVersionPlainText WalletEncryptedVersion = iota
	WalletEncryptedVersionEncryption
)

func (e WalletEncryptedVersion) String() string {
	switch e {
	case WalletEncryptedVersionPlainText:
		return "WalletEncryptedVersionPlainText"
	case WalletEncryptedVersionEncryption:
		return "WalletEncryptedVersionEncryption"
	default:
		return "Unknown EncryptedVersion"
	}
}
