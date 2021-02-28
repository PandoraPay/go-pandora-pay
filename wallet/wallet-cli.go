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

func initWalletCLI() {

	gui.CommandDefineCallback("List Addresses", cliListAddresses)
	gui.CommandDefineCallback("Create New Address", cliCreateNewAddress)
	gui.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic)
	gui.CommandDefineCallback("Show Private Key", cliShowPrivateKey)
	gui.CommandDefineCallback("Remove Address", cliRemoveAddress)
	gui.CommandDefineCallback("Export (JSON) Wallet", cliExportJSONWallet)

}

func cliExportJSONWallet(cmd string) (err error) {

	var f *os.File

	str := <-gui.OutputReadString("Path to export")
	if f, err = os.Create(str); err != nil {
		err = errors.New("File can not be written")
		return
	}
	defer f.Close()

	if err = cliListAddresses(""); err != nil {
		return
	}
	index := <-gui.OutputReadInt("Select Address to be Exported")
	W.RLock()
	defer W.RUnlock()

	if index < 0 {
		err = errors.New("Invalid index")
		return
	}

	var marshal []byte
	var obj interface{}

	if index > len(W.Addresses) {
		obj = W
	} else {
		obj = W.Addresses[index]
	}

	if marshal, err = json.Marshal(obj); err != nil {
		gui.Error("Error marshaling wallet", err)
	}

	if _, err = fmt.Fprint(f, string(marshal)); err != nil {
		err = errors.New("Error writing into file")
		return
	}

	gui.Info("Exported successfully")
	return
}

func cliCreateNewAddress(cmd string) (err error) {

	if err = W.addNewAddress(); err == nil {
		err = cliListAddresses(cmd)
		return
	}

	return
}

func cliRemoveAddress(cmd string) (err error) {

	if err = cliListAddresses(""); err != nil {
		return
	}

	index := <-gui.OutputReadInt("Select Address to be Removed")

	if err = W.removeAddress(index); err == nil {
		_ = cliListAddresses("")
		gui.OutputWrite("Address removed")
	}

	return
}

func cliListAddresses(cmd string) (err error) {

	gui.OutputWrite("Wallet")
	gui.OutputWrite("Version: " + W.Version.String())
	gui.OutputWrite("Encrypted: " + wSaved.Encrypted.String())
	gui.OutputWrite("Count: " + strconv.Itoa(W.Count))

	gui.OutputWrite("")

	err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		var accs *accounts.Accounts
		accs, err = accounts.NewAccounts(tx)

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

func cliShowMnemonic(string) (err error) {
	gui.OutputWrite("Mnemonic \n")
	gui.OutputWrite(W.Mnemonic)

	gui.OutputWrite("Seed \n")
	gui.OutputWrite(W.Seed)

	return
}

func cliShowPrivateKey(cmd string) (err error) {

	if err = cliListAddresses(""); err != nil {
		return
	}

	index := <-gui.OutputReadInt("Select Address")

	var key *[32]byte
	if key, err = W.showPrivateKey(index); err == nil {
		gui.OutputWrite(*key)
	}

	return
}
