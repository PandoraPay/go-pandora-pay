package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/config/config_coins"
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

func (wallet *Wallet) CreateEmptyWallet() (err error) {

	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	wallet.clearWallet()
	wallet.setLoaded(true)

	if err = wallet.createSeed(false); err != nil {
		return
	}
	if _, err = wallet.AddNewAddress(false, "", false, false, true); err != nil {
		return
	}

	return
}

func (wallet *Wallet) ImportMnemonic(mnemonic string) (err error) {

	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	wallet.clearWallet()
	wallet.setLoaded(true)

	wallet.Mnemonic = mnemonic

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "SEED Secret Passphrase")
	if err != nil {
		return err
	}

	wallet.Seed = seed

	if _, err = wallet.AddNewAddress(false, "", false, false, true); err != nil {
		return
	}

	return
}

func (wallet *Wallet) updateWallet() {
	gui.GUI.InfoUpdate("Wallet Addrs", fmt.Sprintf("%d  %s", wallet.Count, wallet.Encryption.Encrypted))
}

//it must be locked and use original walletAddresses, not cloned ones
func (wallet *Wallet) refreshWalletAccount(acc *account.Account, reg *registration.Registration, chainHeight uint64, addr *wallet_address.WalletAddress) (err error) {

	deleted := false

	if acc == nil || reg == nil || !reg.Staked || addr.SharedStaked == nil {
		deleted = true
	} else {

		stakingAmountBalance := acc.Balance.Amount.Serialize()

		var stakingAmount uint64
		if stakingAmountBalance != nil {
			stakingAmount, _ = wallet.DecryptBalance(addr, stakingAmountBalance, config_coins.NATIVE_ASSET_FULL, false, 0, true, context.Background(), func(string) {})
		}

		if stakingAmount < config_stake.GetRequiredStake(chainHeight) {
			deleted = true
		}

	}

	if deleted {

		wallet.forging.Wallet.RemoveWallet(addr.PublicKey, true, acc, reg, chainHeight)

		if addr.IsSharedStaked {
			_, err = wallet.RemoveAddressByPublicKey(addr.PublicKey, true)
			return
		}

	} else {
		wallet.forging.Wallet.AddWallet(addr.PublicKey, addr.SharedStaked, true, acc, reg, chainHeight)
	}

	return
}
