package wallet

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/context"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	wallet_address "pandora-pay/wallet/address"
	"strconv"
)

func (wallet *Wallet) CliListAddresses(cmd string) (err error) {

	context.GUI.OutputWrite("Wallet")
	context.GUI.OutputWrite("Version: " + wallet.Version.String())
	context.GUI.OutputWrite("Encrypted: " + wallet.Encrypted.String())
	context.GUI.OutputWrite("Count: " + strconv.Itoa(wallet.Count))

	context.GUI.OutputWrite("")

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get([]byte("chainHeight")))

		accs := accounts.NewAccounts(reader)
		toks := tokens.NewTokens(reader)

		for _, walletAddress := range wallet.Addresses {
			addressStr := walletAddress.GetAddressEncoded()
			context.GUI.OutputWrite(walletAddress.Name + " : " + walletAddress.Address.Version.String() + " : " + addressStr)

			if walletAddress.Address.Version == addresses.SimplePublicKeyHash ||
				walletAddress.Address.Version == addresses.SimplePublicKey {

				var acc *account.Account
				if acc, err = accs.GetAccount(walletAddress.GetPublicKeyHash(), chainHeight); err != nil {
					return
				}

				if acc == nil {
					context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
				} else {
					context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Nonce", strconv.FormatUint(acc.Nonce, 10)))
					if len(acc.Balances) > 0 {
						context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", ""))
						for _, balance := range acc.Balances {

							var tok *token.Token
							if tok, err = toks.GetToken(balance.Token); err != nil {
								return
							}
							context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", strconv.FormatFloat(config.ConvertToBase(balance.Amount), 'f', config.DECIMAL_SEPARATOR, 64), tok.Name))
						}
					} else {
						context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", "EMPTY"))
					}
					if acc.HasDelegatedStake() {
						context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Stake Available", strconv.FormatFloat(config.ConvertToBase(acc.DelegatedStake.StakeAvailable), 'f', config.DECIMAL_SEPARATOR, 64)))

						if len(acc.DelegatedStake.StakesPending) > 0 {
							context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES", ""))
							for _, stakePending := range acc.DelegatedStake.StakesPending {
								context.GUI.OutputWrite(fmt.Sprintf("%18s: %10s %t", strconv.FormatUint(stakePending.ActivationHeight, 10), strconv.FormatFloat(config.ConvertToBase(stakePending.PendingAmount), 'f', config.DECIMAL_SEPARATOR, 64), stakePending.PendingType))
							}
						} else {
							context.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES:", "EMPTY"))
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

	index, ok := context.GUI.OutputReadInt(text, nil)
	if !ok {
		err = errors.New("Canceled")
		return
	}

	walletAddress, err = wallet.GetWalletAddress(index)
	return
}

func (wallet *Wallet) initWalletCLI() {

	cliExportAddressJSON := func(cmd string) (err error) {

		str, ok := context.GUI.OutputReadString("Path to export")
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
		index, ok := context.GUI.OutputReadInt("Select Address to be Exported", nil)
		if !ok {
			return
		}

		wallet.RLock()
		defer wallet.RUnlock()

		if index < 0 {
			return errors.New("Invalid index")
		}
		if index >= len(wallet.Addresses) {
			return errors.New("Address index is invalid")
		}

		obj := wallet.Addresses[index]

		var marshal []byte
		if marshal, err = json.Marshal(obj); err != nil {
			return errors.New("Error marshaling wallet")
		}

		if _, err = fmt.Fprint(f, string(marshal)); err != nil {
			return errors.New("Error writing into file")
		}

		context.GUI.Info("Exported successfully")
		return
	}

	cliImportAddressJSON := func(cmd string) (err error) {

		str, ok := context.GUI.OutputReadString("Path to import")
		if !ok {
			return
		}

		data, err := os.ReadFile(str)
		if err != nil {
			return
		}

		wallet.RLock()
		defer wallet.RUnlock()

		adr := &wallet_address.WalletAddress{}

		if err = json.Unmarshal(data, adr); err != nil {
			return errors.New("Error unmarshaling wallet")
		}

		isMine := false
		if wallet.SeedIndex != 0 {
			key, err := wallet.GeneratePrivateKey(adr.SeedIndex, false)
			if err == nil && key != nil {
				isMine = true
			}
		}

		if !isMine {
			adr.IsMine = false
			adr.SeedIndex = 0
		}

		if err = wallet.AddAddress(adr, false, false, isMine); err != nil {
			return
		}

		context.GUI.Info("Imported successfully")
		return
	}

	cliExportWalletJSON := func(cmd string) (err error) {

		str, ok := context.GUI.OutputReadString("Path to export")
		if !ok {
			return
		}

		f, err := os.Create(str)
		if err != nil {
			return
		}

		defer f.Close()

		wallet.RLock()
		defer wallet.RUnlock()

		var marshal []byte
		if marshal, err = json.Marshal(wallet); err != nil {
			return errors.New("Error marshaling wallet")
		}

		if _, err = fmt.Fprint(f, string(marshal)); err != nil {
			return errors.New("Error writing into file")
		}

		context.GUI.Info("Wallet Exported successfully")
		return
	}

	cliImportWalletJSON := func(cmd string) (err error) {

		str, ok := context.GUI.OutputReadString("Path to import Wallet")
		if !ok {
			return
		}

		done, ok := context.GUI.OutputReadBool("Your wallet will be REPLACED with this one. Are you sure ?")
		if !ok {
			return
		}

		if !done {
			return errors.New("You didn't accept REPLACING your existing wallet")
		}

		data, err := os.ReadFile(str)
		if err != nil {
			return
		}

		wallet2 := createWallet(wallet.forging, wallet.mempool)
		if err = json.Unmarshal(data, wallet2); err != nil {
			return errors.New("Error unmarshaling wallet")
		}

		wallet.RLock()
		defer wallet.RUnlock()

		if err = json.Unmarshal(data, wallet); err != nil {
			return errors.New("Error unmarshaling wallet 2")
		}

		wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)
		for _, adr := range wallet.Addresses {
			wallet.addressesMap[string(adr.Address.PublicKeyHash)] = adr
		}

		context.GUI.Info("Wallet Imported Successfully")
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
			context.GUI.OutputWrite("Address removed")
		} else {
			context.GUI.OutputWrite("Address was NOT removed ")
		}
		return
	}

	cliShowMnemonic := func(string) (err error) {
		context.GUI.OutputWrite("Mnemonic \n")
		context.GUI.OutputWrite(wallet.Mnemonic)

		context.GUI.OutputWrite("Seed \n")
		context.GUI.OutputWrite(wallet.Seed)

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
		context.GUI.OutputWrite(privateKey)

		return
	}

	cliImportPrivateKey := func(cmd string) (err error) {

		privateKey, ok := context.GUI.OutputReadBytes("Write Private key", []int{32})
		if !ok {
			return
		}

		name, ok := context.GUI.OutputReadString("Write Name of the newly imported address")
		if !ok {
			return
		}

		var adr *wallet_address.WalletAddress
		if adr, err = wallet.ImportPrivateKey(name, privateKey); err != nil {
			return
		}

		context.GUI.OutputWrite("Address was imported: " + adr.AddressEncoded)

		return
	}

	context.GUI.CommandDefineCallback("List Addresses", wallet.CliListAddresses)
	context.GUI.CommandDefineCallback("Create New Address", cliCreateNewAddress)
	context.GUI.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic)
	context.GUI.CommandDefineCallback("Show Private Key", cliShowPrivateKey)
	context.GUI.CommandDefineCallback("Import Private Key", cliImportPrivateKey)
	context.GUI.CommandDefineCallback("Remove Address", cliRemoveAddress)
	context.GUI.CommandDefineCallback("Export Address JSON", cliExportAddressJSON)
	context.GUI.CommandDefineCallback("Import Address JSON", cliImportAddressJSON)
	context.GUI.CommandDefineCallback("Export Wallet JSON", cliExportWalletJSON)
	context.GUI.CommandDefineCallback("Import Wallet JSON", cliImportWalletJSON)

}
