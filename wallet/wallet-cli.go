package wallet

import (
	"encoding/hex"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/account"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/store"
	"strconv"
)

func initWalletCLI() {

	gui.CommandDefineCallback("List Addresses", cliListAddresses)
	gui.CommandDefineCallback("Create New Address", cliCreateNewAddress)
	gui.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic)
	gui.CommandDefineCallback("Show Private Key", cliShowPrivateKey)
	gui.CommandDefineCallback("Remove Address", cliRemoveAddress)

}

func cliCreateNewAddress(cmd string) {

	if err := W.addNewAddress(); err != nil {
		gui.Error("Error creating a new Address", err)
	}

	cliListAddresses(cmd)
}

func cliRemoveAddress(cmd string) {
	cliListAddresses("")
	index := <-gui.OutputReadInt("Select Address to be Removed")

	if err := W.removeAddress(index); err != nil {
		gui.Error(err)
	} else {
		cliListAddresses("")
		gui.OutputWrite("Address removed")
	}
	gui.OutputDone()
}

func cliListAddresses(cmd string) {

	gui.OutputWrite("Wallet")
	gui.OutputWrite("Version: " + W.Version.String())
	gui.OutputWrite("Encrypted: " + wSaved.Encrypted.String())
	gui.OutputWrite("Count: " + strconv.Itoa(W.Count))

	gui.OutputWrite("")

	err := store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		var accs *accounts.Accounts
		accs, err = accounts.CreateNewAccounts(tx)

		for _, walletAddress := range W.Addresses {
			addressStr, _ := walletAddress.Address.EncodeAddr()
			gui.OutputWrite(walletAddress.Name + " : " + walletAddress.Address.Version.String() + " : " + addressStr)

			if walletAddress.Address.Version == addresses.TransparentPublicKeyHash ||
				walletAddress.Address.Version == addresses.TransparentPublicKey {

				var acc *account.Account
				if acc, err = accs.GetAccount(walletAddress.PublicKeyHash); err != nil {
					return
				}

				if acc == nil {
					gui.OutputWrite("      -> " + "EMPTY")
				} else {
					for _, balance := range acc.Balances {
						gui.OutputWrite("      -> " + strconv.FormatUint(config.ConvertToBase(balance.Amount), 10) + " " + hex.EncodeToString(balance.Currency))
					}
				}

			}

		}

		return
	})

	if err != nil {
		gui.Error(err)
	}

	if cmd != "" {
		gui.OutputDone()
	}
}

func cliShowMnemonic(string) {
	gui.OutputWrite("Mnemonic \n")
	gui.OutputWrite(W.Mnemonic)

	gui.OutputWrite("Seed \n")
	gui.OutputWrite(W.Seed)
	gui.OutputDone()
}

func cliShowPrivateKey(cmd string) {

	cliListAddresses("")

	index := <-gui.OutputReadInt("Select Address")

	if key, err := W.showPrivateKey(index); err != nil {
		gui.Error(err)
	} else {
		gui.OutputWrite(*key)
	}

	gui.OutputDone()
}
