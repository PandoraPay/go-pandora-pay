package wallet_address

type WalletAddressDecryptedBalance struct {
	Amount           uint64 `json:"amount" msgpack:"amount"`
	EncryptedBalance []byte `json:"encryptedBalance" msgpack:"encryptedBalance"`
}
