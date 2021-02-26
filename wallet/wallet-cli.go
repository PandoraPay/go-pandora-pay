package wallet

import (
	"encoding/hex"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/account"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/config"
	"pandora-pay/crypto"
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
	err := addNewAddress()
	if err != nil {
		gui.Error("Error creating a new Address", err)
	}
	cliListAddresses(cmd)
}

func cliRemoveAddress(cmd string) {
	cliListAddresses("")
	index := <-gui.OutputReadInt("Select Address to be Removed")

	err := removeAddress(index)
	if err != nil {
		gui.Error(err)
	} else {
		cliListAddresses("")
		gui.OutputWrite("Address removed")
	}
	gui.OutputDone()
}

func cliListAddresses(cmd string) {

	gui.OutputWrite("Wallet")
	gui.OutputWrite("Version: " + wallet.Version.String())
	gui.OutputWrite("Encrypted: " + walletSaved.Encrypted.String())
	gui.OutputWrite("Count: " + strconv.Itoa(wallet.Count))

	gui.OutputWrite("")

	err := store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		var accs *accounts.Accounts
		accs, err = accounts.CreateNewAccounts(tx, true)

		for _, walletAddress := range wallet.Addresses {
			addressStr, _ := walletAddress.Address.EncodeAddr()
			gui.OutputWrite(walletAddress.Name + " : " + walletAddress.Address.Version.String() + " : " + addressStr)

			if walletAddress.Address.Version == addresses.TransparentPublicKeyHash ||
				walletAddress.Address.Version == addresses.TransparentPublicKey {

				publicKeyHash := walletAddress.Address.PublicKey
				if walletAddress.Address.Version == addresses.TransparentPublicKey {
					publicKeyHash = crypto.ComputePublicKeyHash(publicKeyHash)
				}

				var finaPublicKeyHash [20]byte
				copy(finaPublicKeyHash[:], publicKeyHash)

				var acc *account.Account
				if acc, err = accs.GetAccount(finaPublicKeyHash, false); err != nil {
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
	gui.OutputWrite(wallet.Mnemonic)

	gui.OutputWrite("Seed \n")
	gui.OutputWrite(wallet.Seed)
	gui.OutputDone()
}

func cliShowPrivateKey(cmd string) {

	cliListAddresses("")

	index := <-gui.OutputReadInt("Select Address")
	key, err := showPrivateKey(index)
	if err != nil {
		gui.Error(err)
	} else {
		gui.OutputWrite(key)
	}

	gui.OutputDone()
}
