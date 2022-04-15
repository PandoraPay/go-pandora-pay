package wallet

import (
	"context"
	"errors"
	"fmt"
	"pandora-pay/addresses"
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

	for {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			continue
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			continue
		}

		wallet.Mnemonic = mnemonic

		// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "SEED Secret Passphrase")
		if err != nil {
			continue
		}

		var seedExtended *addresses.SeedExtended
		if seedExtended, err = addresses.NewSeedExtended(seed); err != nil {
			continue
		}

		wallet.Seed = seedExtended.Serialize()
		return nil
	}
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
		return
	}

	seedExtended, err := addresses.NewSeedExtended(seed)
	if err != nil {
		return
	}

	wallet.Seed = seedExtended.Serialize()

	if _, err = wallet.AddNewAddress(false, "", false, false, true); err != nil {
		return
	}

	return
}

func (wallet *Wallet) ImportEntropy(entropy []byte) (err error) {

	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	wallet.clearWallet()
	wallet.setLoaded(true)

	if wallet.Mnemonic, err = bip39.NewMnemonic(entropy); err != nil {
		return
	}

	seed, err := bip39.NewSeedWithErrorChecking(wallet.Mnemonic, "SEED Secret Passphrase")
	if err != nil {
		return err
	}

	seedExtended, err := addresses.NewSeedExtended(seed)
	if err != nil {
		return
	}

	wallet.Seed = seedExtended.Serialize()

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
