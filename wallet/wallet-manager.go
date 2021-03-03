package wallet

import (
	"encoding/json"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/forging"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
	"sync"
)

type EncryptedVersion int

const (
	PlainText EncryptedVersion = iota
	Encrypted
)

func (e EncryptedVersion) String() string {
	switch e {
	case PlainText:
		return "PlainText"
	case Encrypted:
		return "Encrypted"
	default:
		return "Unknown EncryptedVersion"
	}
}

type Wallet struct {
	Encrypted EncryptedVersion

	Version   Version
	Mnemonic  string
	Seed      [32]byte
	SeedIndex uint32
	Count     int

	Addresses []*WalletAddress `json:"-"` //stored separately
	Checksum  [4]byte          `json:"-"`

	// forging creates multiple threads and it will read the wallet.Addresses
	sync.RWMutex `json:"-"`
}

func (wallet *Wallet) addNewAddress() (err error) {

	masterKey, _ := bip32.NewMasterKey(wallet.Seed[:])

	var key *bip32.Key
	if key, err = masterKey.NewChildKey(wallet.SeedIndex); err != nil {
		gui.Fatal("Couldn't derivate the marker key", err)
	}

	privateKey := addresses.PrivateKey{Key: *helpers.Byte32(key.Key)}

	var publicKey [33]byte
	if publicKey, err = privateKey.GeneratePublicKey(); err != nil {
		gui.Fatal("Generating Public Key from Private key raised an error", err)
	}

	var address *addresses.Address
	if address, err = privateKey.GenerateAddress(true, 0, []byte{}); err != nil {
		gui.Fatal("Generating Address raised an error", err)
	}

	publicKeyHash := crypto.ComputePublicKeyHash(publicKey)

	wallet.Lock()
	defer wallet.Unlock()
	walletAddress := WalletAddress{
		"Addr " + strconv.Itoa(wallet.Count),
		&privateKey,
		publicKey,
		publicKeyHash,
		address,
		wallet.SeedIndex,
	}

	wallet.Addresses = append(wallet.Addresses, &walletAddress)
	wallet.Count += 1
	wallet.SeedIndex += 1

	go forging.ForgingW.AddWallet(publicKey, privateKey.Key, publicKeyHash)

	wallet.updateWallet()
	return wallet.saveWallet()
}

func (wallet *Wallet) removeAddress(index int) error {

	wallet.Lock()
	defer wallet.Unlock()

	if index < 0 || index > len(wallet.Addresses) {
		return errors.New("Invalid Address Index")
	}

	removing := wallet.Addresses[index]

	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	wallet.Count -= 1

	go forging.ForgingW.RemoveWallet(removing.PublicKey)

	wallet.updateWallet()
	return wallet.saveWallet()
}

func (wallet *Wallet) showPrivateKey(index int) (*[32]byte, error) {

	wallet.RLock()
	defer wallet.RUnlock()

	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return &wallet.Addresses[index].PrivateKey.Key, nil
}

func (wallet *Wallet) createSeed() (err error) {

	wallet.Lock()
	defer wallet.Unlock()

	var entropy []byte
	if entropy, err = bip39.NewEntropy(256); err != nil {
		return gui.Error("Entropy of the address raised an error", err)
	}

	var mnemonic string
	if mnemonic, err = bip39.NewMnemonic(entropy); err != nil {
		return gui.Error("Mnemonic couldn't be created", err)
	}

	wallet.Mnemonic = mnemonic

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "SEED Secret Passphrase")
	wallet.Seed = *helpers.Byte32(seed)

	return nil
}

func (wallet *Wallet) createEmptyWallet() error {
	if err := wallet.createSeed(); err != nil {
		return gui.Error("Error creating seed", err)
	}
	return wallet.addNewAddress()
}

func (wallet *Wallet) updateWallet() {
	gui.InfoUpdate("Wallet", wallet.Encrypted.String())
	gui.InfoUpdate("Wallet Addrs", strconv.Itoa(wallet.Count))
}

func (wallet *Wallet) computeChecksum() (checksum [4]byte, err error) {

	writer := helpers.NewBufferWriter()

	var data []byte
	if data, err = json.Marshal(wallet); err != nil {
		return
	}
	writer.Write(data)

	for _, walletAddress := range wallet.Addresses {
		data, err = json.Marshal(walletAddress)
		writer.Write(data)
	}

	data = writer.Bytes()
	out := crypto.RIPEMD(data)[0:helpers.ChecksumSize]
	copy(checksum[:], out[:])

	return
}
