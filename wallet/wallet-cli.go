package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"os"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/store"
	wallet_address "pandora-pay/wallet/address"
	"strconv"
)

func (wallet *Wallet) CliListAddresses(cmd string) (err error) {

	gui.OutputWrite("Wallet")
	gui.OutputWrite("Version: " + wallet.Version.String())
	gui.OutputWrite("Encrypted: " + wallet.Encrypted.String())
	gui.OutputWrite("Count: " + strconv.Itoa(wallet.Count))

	gui.OutputWrite("")

	return store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

		accs := accounts.NewAccounts(boltTx)
		toks := tokens.NewTokens(boltTx)

		for _, walletAddress := range wallet.Addresses {
			addressStr := walletAddress.GetAddressEncoded()
			gui.OutputWrite(walletAddress.Name + " : " + walletAddress.Address.Version.String() + " : " + addressStr)

			if walletAddress.Address.Version == addresses.SimplePublicKeyHash ||
				walletAddress.Address.Version == addresses.SimplePublicKey {

				acc := accs.GetAccount(walletAddress.GetPublicKeyHash())

				gui.OutputWrite(fmt.Sprintf("%18s: %s", "Nonce", strconv.FormatUint(acc.Nonce, 10)))
				if acc == nil {
					gui.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
				} else {
					if len(acc.Balances) > 0 {
						gui.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", ""))
						for _, balance := range acc.Balances {

							token := toks.GetToken(balance.Token)
							gui.OutputWrite(fmt.Sprintf("%18s: %s", strconv.FormatFloat(config.ConvertToBase(balance.Amount), 'f', config.DECIMAL_SEPARATOR, 64), token.Name))
						}
					} else {
						gui.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", "EMPTY"))
					}
					if acc.HasDelegatedStake() {
						gui.OutputWrite(fmt.Sprintf("%18s: %s", "Stake Available", strconv.FormatFloat(config.ConvertToBase(acc.DelegatedStake.StakeAvailable), 'f', config.DECIMAL_SEPARATOR, 64)))

						if len(acc.DelegatedStake.StakesPending) > 0 {
							gui.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES", ""))
							for _, stakePending := range acc.DelegatedStake.StakesPending {
								gui.OutputWrite(fmt.Sprintf("%18s: %10s %t", strconv.FormatUint(stakePending.ActivationHeight, 10), strconv.FormatFloat(config.ConvertToBase(stakePending.PendingAmount), 'f', config.DECIMAL_SEPARATOR, 64), stakePending.PendingType))
							}
						} else {
							gui.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES:", "EMPTY"))
						}
					}
				}

			}

		}

		return
	})
}

func (wallet *Wallet) CliSelectAddress(text string) (walletAddress *wallet_address.WalletAddress, index int, err error) {

	if err = wallet.CliListAddresses(""); err != nil {
		return
	}

	index, ok := gui.OutputReadInt(text, nil)
	if !ok {
		return
	}

	walletAddress, err = wallet.GetWalletAddress(index)
	return
}

func (wallet *Wallet) initWalletCLI() {

	cliExportJSONWallet := func(cmd string) (err error) {

		str, ok := gui.OutputReadString("Path to export")
		if !ok {
			return
		}
		f, err := os.Create(str)
		if err != nil {
			return
		}

		defer f.Close()

		if err = wallet.CliListAddresses(""); err != nil {
			return
		}
		index, ok := gui.OutputReadInt("Select Address to be Exported", nil)
		if !ok {
			return
		}

		wallet.RLock()
		defer wallet.RUnlock()

		if index < 0 {
			return errors.New("Invalid index")
		}

		var marshal []byte
		var obj interface{}

		if index > len(wallet.Addresses) {
			obj = wallet
		} else {
			obj = wallet.Addresses[index]
		}

		if marshal, err = json.Marshal(obj); err != nil {
			return errors.New("Error marshaling wallet")
		}

		if _, err = fmt.Fprint(f, string(marshal)); err != nil {
			return errors.New("Error writing into file")
		}

		gui.Info("Exported successfully")
		return
	}

	cliCreateNewAddress := func(cmd string) (err error) {
		if _, err = wallet.AddNewAddress(); err != nil {
			return
		}
		return wallet.CliListAddresses(cmd)
	}

	cliRemoveAddress := func(cmd string) (err error) {

		_, index, err := wallet.CliSelectAddress("Select Address to be Removed")
		if err != nil {
			return
		}

		var success bool
		if success, err = wallet.RemoveAddress(index); err != nil {
			return
		}
		if err = wallet.CliListAddresses(""); err != nil {
			return
		}

		if success {
			gui.OutputWrite("Address removed")
		} else {
			gui.OutputWrite("Address was NOT removed ")
		}
		return
	}

	cliShowMnemonic := func(string) (err error) {
		gui.OutputWrite("Mnemonic \n")
		gui.OutputWrite(wallet.Mnemonic)

		gui.OutputWrite("Seed \n")
		gui.OutputWrite(wallet.Seed)

		return
	}

	cliShowPrivateKey := func(cmd string) (err error) {

		_, index, err := wallet.CliSelectAddress("Select Address to be Removed")
		if err != nil {
			return
		}

		privateKey, err := wallet.ShowPrivateKey(index)
		if err != nil {
			return
		}
		gui.OutputWrite(privateKey)

		return
	}

	gui.CommandDefineCallback("Wallet  : List Addresses", wallet.CliListAddresses)
	gui.CommandDefineCallback("Wallet  : Create New Address", cliCreateNewAddress)
	gui.CommandDefineCallback("Wallet  : Show Mnemnonic", cliShowMnemonic)
	gui.CommandDefineCallback("Wallet  : Show Private Key", cliShowPrivateKey)
	gui.CommandDefineCallback("Wallet  : Remove Address", cliRemoveAddress)
	gui.CommandDefineCallback("Wallet  : Export JSON", cliExportJSONWallet)

}
