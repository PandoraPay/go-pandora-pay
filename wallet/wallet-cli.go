package wallet

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"os"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/store"
	"strconv"
)

func initWalletCLI(wallet *Wallet) {

	cliListAddresses := func(cmd string) (err error) {

		gui.OutputWrite("Wallet")
		gui.OutputWrite("Version: " + wallet.Version.String())
		gui.OutputWrite("Encrypted: " + wallet.Encrypted.String())
		gui.OutputWrite("Count: " + strconv.Itoa(wallet.Count))

		gui.OutputWrite("")

		err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

			var accs *accounts.Accounts
			accs, err = accounts.NewAccounts(tx)

			for _, walletAddress := range wallet.Addresses {
				addressStr, _ := walletAddress.Address.EncodeAddr()
				gui.OutputWrite(walletAddress.Name + " : " + walletAddress.Address.Version.String() + " : " + addressStr)

				if walletAddress.Address.Version == addresses.SimplePublicKeyHash ||
					walletAddress.Address.Version == addresses.SimplePublicKey {

					var acc *account.Account
					if acc, err = accs.GetAccount(walletAddress.PublicKeyHash); err != nil {
						return
					}

					if acc == nil {
						gui.OutputWrite("      -> " + "EMPTY")
					} else {
						for _, balance := range acc.Balances {
							gui.OutputWrite("      -> " + strconv.FormatUint(config.ConvertToBase(balance.Amount), 10) + " " + hex.EncodeToString(balance.Token))
						}
						if acc.HasDelegatedStake() {
							gui.OutputWrite("      ->   Stake Available   " + strconv.FormatUint(config.ConvertToBase(acc.DelegatedStake.StakeAvailable), 10))
							gui.OutputWrite("      ->   Unstake Available " + strconv.FormatUint(config.ConvertToBase(acc.DelegatedStake.UnstakeAmount), 10))
						}
					}

				}

			}

			return
		})

		return
	}

	cliExportJSONWallet := func(cmd string) (err error) {

		var f *os.File

		str := <-gui.OutputReadString("Path to export")
		if f, err = os.Create(str); err != nil {
			return errors.New("File can not be written")
		}
		defer f.Close()

		if err = cliListAddresses(""); err != nil {
			return
		}
		index := <-gui.OutputReadInt("Select Address to be Exported")
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
			gui.Error("Error marshaling wallet", err)
		}

		if _, err = fmt.Fprint(f, string(marshal)); err != nil {
			return errors.New("Error writing into file")
		}

		gui.Info("Exported successfully")
		return
	}

	cliCreateNewAddress := func(cmd string) (err error) {

		if err = wallet.addNewAddress(); err == nil {
			err = cliListAddresses(cmd)
			return
		}

		return
	}

	cliRemoveAddress := func(cmd string) (err error) {

		if err = cliListAddresses(""); err != nil {
			return
		}

		index := <-gui.OutputReadInt("Select Address to be Removed")

		if err = wallet.removeAddress(index); err == nil {
			_ = cliListAddresses("")
			gui.OutputWrite("Address removed")
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

		if err = cliListAddresses(""); err != nil {
			return
		}

		index := <-gui.OutputReadInt("Select Address")

		var key *[32]byte
		if key, err = wallet.showPrivateKey(index); err == nil {
			gui.OutputWrite(*key)
		}

		return
	}

	gui.CommandDefineCallback("List Addresses", cliListAddresses)
	gui.CommandDefineCallback("Create New Address", cliCreateNewAddress)
	gui.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic)
	gui.CommandDefineCallback("Show Private Key", cliShowPrivateKey)
	gui.CommandDefineCallback("Remove Address", cliRemoveAddress)
	gui.CommandDefineCallback("Export (JSON) Wallet", cliExportJSONWallet)

}
