package wallet

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_nodes"
	"pandora-pay/gui"
	"pandora-pay/wallet/wallet_address"
)

func (wallet *Wallet) createSeed(lock bool) error {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return errors.New("Entropy of the address raised an error")
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return errors.New("Mnemonic couldn't be created")
	}

	wallet.Mnemonic = mnemonic

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "SEED Secret Passphrase")
	wallet.Seed = seed
	return nil
}

func (wallet *Wallet) createEmptyWallet() (err error) {
	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	wallet.setLoaded(true)
	if err = wallet.createSeed(false); err != nil {
		return
	}
	if _, err = wallet.AddNewAddress(false); err != nil {
		return
	}

	return
}

func (wallet *Wallet) updateWallet() {
	gui.GUI.InfoUpdate("Wallet Addrs", fmt.Sprintf("%d  %s", wallet.Count, wallet.Encryption.Encrypted))
}

func (wallet *Wallet) refreshWalletPlainAccount(plainAcc *plain_account.PlainAccount, addr *wallet_address.WalletAddress, lock bool) (err error) {

	if plainAcc == nil {
		return
	}

	if addr.DelegatedStake != nil && !plainAcc.DelegatedStake.HasDelegatedStake() {
		addr.DelegatedStake = nil

		if addr.PrivateKey == nil {
			_, err = wallet.RemoveAddressByPublicKey(addr.PublicKey, lock)
			return
		}

		return
	}

	if (addr.DelegatedStake != nil && plainAcc.DelegatedStake.HasDelegatedStake() && !bytes.Equal(addr.DelegatedStake.PublicKey, plainAcc.DelegatedStake.DelegatedStakePublicKey)) ||
		(addr.DelegatedStake == nil && plainAcc.DelegatedStake.HasDelegatedStake()) {

		if addr.PrivateKey == nil {
			_, err = wallet.RemoveAddressByPublicKey(addr.PublicKey, lock)
			return
		}

		if plainAcc.DelegatedStake.HasDelegatedStake() {

			if plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
				_, err = wallet.RemoveAddressByPublicKey(addr.PublicKey, lock)
				return
			}

			lastKnownNonce := uint32(0)
			if addr.DelegatedStake != nil {
				lastKnownNonce = addr.DelegatedStake.LastKnownNonce
			}

			var delegatedStake *wallet_address.WalletAddressDelegatedStake
			if delegatedStake, err = addr.FindDelegatedStake(uint32(plainAcc.Nonce), lastKnownNonce, plainAcc.DelegatedStake.DelegatedStakePublicKey); err != nil {
				_, err = wallet.RemoveAddressByPublicKey(addr.PublicKey, lock)
				return
			}

			if delegatedStake != nil {
				addr.DelegatedStake = delegatedStake
				wallet.forging.Wallet.AddWallet(addr.DelegatedStake.PrivateKey.Key, addr.PublicKey, true, plainAcc)
				return wallet.saveWalletAddress(addr, lock)
			}

		}

		addr.DelegatedStake = nil
		wallet.forging.Wallet.RemoveWallet(addr.PublicKey, true, plainAcc)

		return wallet.saveWalletAddress(addr, lock)
	}

	return
}
