package wallet

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"math"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/config_stake"
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
	if _, err = wallet.AddNewAddress(false, ""); err != nil {
		return
	}

	return
}

func (wallet *Wallet) updateWallet() {
	gui.GUI.InfoUpdate("Wallet Addrs", fmt.Sprintf("%d  %s", wallet.Count, wallet.Encryption.Encrypted))
}

//it must be locked and use original walletAddresses, not cloned ones
func (wallet *Wallet) refreshWalletPlainAccount(plainAcc *plain_account.PlainAccount, chainHeight uint64, addr *wallet_address.WalletAddress, lock bool) (err error) {

	if lock {
		return errors.New("wallet should be locked before")
	}

	prevDelegatedStake := addr.DelegatedStake

	if plainAcc == nil || !plainAcc.DelegatedStake.HasDelegatedStake() {
		addr.DelegatedStake = nil
	} else {
		if plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
			addr.DelegatedStake = nil
		} else if addr.DelegatedStake != nil && !bytes.Equal(addr.DelegatedStake.PublicKey, plainAcc.DelegatedStake.DelegatedStakePublicKey) {
			addr.DelegatedStake = nil
		} else {
			var amount uint64
			if amount, err = plainAcc.DelegatedStake.ComputeDelegatedStakeAvailable(math.MaxUint64); err != nil {
				addr.DelegatedStake = nil
			}
			if amount < config_stake.GetRequiredStake(chainHeight) {
				addr.DelegatedStake = nil
			}
		}
	}

	if addr.DelegatedStake == nil {

		if addr.PrivateKey == nil {
			wallet.forging.Wallet.RemoveWallet(addr.PublicKey, true, plainAcc, chainHeight)
			_, err = wallet.RemoveAddressByPublicKey(addr.PublicKey, lock)
			return
		}

		lastKnownNonce := uint32(0)
		if addr.DelegatedStake != nil {
			lastKnownNonce = addr.DelegatedStake.LastKnownNonce
		}

		if plainAcc != nil {
			var delegatedStake *wallet_address.WalletAddressDelegatedStake
			if delegatedStake, err = addr.FindDelegatedStake(uint32(plainAcc.Nonce), lastKnownNonce, plainAcc.DelegatedStake.DelegatedStakePublicKey); err != nil {
				_, err = wallet.RemoveAddressByPublicKey(addr.PublicKey, lock)
				return
			}

			if delegatedStake != nil {
				addr.DelegatedStake = delegatedStake
				wallet.forging.Wallet.AddWallet(addr.DelegatedStake.PrivateKey.Key, addr.PublicKey, true, plainAcc, chainHeight)
			}
		}

	}

	if prevDelegatedStake != addr.DelegatedStake {
		return wallet.saveWalletAddress(addr, lock)
	}

	return
}
