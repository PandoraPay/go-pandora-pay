package wallet

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	wallet_address "pandora-pay/wallet/address"
	"strconv"
)

func (wallet *Wallet) CliListAddresses(cmd string) (err error) {

	gui.GUI.OutputWrite("Wallet")
	gui.GUI.OutputWrite("Version: " + wallet.Version.String())
	gui.GUI.OutputWrite("Encrypted: " + wallet.Encryption.Encrypted.String())
	gui.GUI.OutputWrite("Count: " + strconv.Itoa(wallet.Count))

	gui.GUI.OutputWrite("")

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)
		toks := tokens.NewTokens(reader)

		for _, walletAddress := range wallet.Addresses {
			addressStr := walletAddress.AddressEncoded
			gui.GUI.OutputWrite(walletAddress.Name + " : " + walletAddress.Version.String() + " : " + addressStr)

			if walletAddress.Version == wallet_address.VERSION_TRANSPARENT {

				var acc *account.Account
				if acc, err = accs.GetAccount(walletAddress.PublicKeyHash, chainHeight); err != nil {
					return
				}

				if acc == nil {
					gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
				} else {
					gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Nonce", strconv.FormatUint(acc.Nonce, 10)))
					if len(acc.Balances) > 0 {
						gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", ""))
						for _, balance := range acc.Balances {

							var tok *token.Token
							if tok, err = toks.GetToken(balance.Token); err != nil {
								return
							}
							gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", strconv.FormatFloat(config.ConvertToBase(balance.Amount), 'f', config.DECIMAL_SEPARATOR, 64), tok.Name))
						}
					} else {
						gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", "EMPTY"))
					}
					if acc.HasDelegatedStake() {
						gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Stake Available", strconv.FormatFloat(config.ConvertToBase(acc.DelegatedStake.StakeAvailable), 'f', config.DECIMAL_SEPARATOR, 64)))

						if len(acc.DelegatedStake.StakesPending) > 0 {
							gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES", ""))
							for _, stakePending := range acc.DelegatedStake.StakesPending {
								gui.GUI.OutputWrite(fmt.Sprintf("%18s: %10s %t", strconv.FormatUint(stakePending.ActivationHeight, 10), strconv.FormatFloat(config.ConvertToBase(stakePending.PendingAmount), 'f', config.DECIMAL_SEPARATOR, 64), stakePending.PendingType))
							}
						} else {
							gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES:", "EMPTY"))
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

	index, ok := gui.GUI.OutputReadInt(text, nil)
	if !ok {
		err = errors.New("Canceled")
		return
	}

	walletAddress, err = wallet.GetWalletAddress(index)
	return
}

func (wallet *Wallet) initWalletCLI() {

	cliExportAddressJSON := func(cmd string) (err error) {

		str, ok := gui.GUI.OutputReadString("Path to export")
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
		index, ok := gui.GUI.OutputReadInt("Select Address to be Exported", nil)
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

		gui.GUI.Info("Exported successfully")
		return
	}

	cliImportAddressJSON := func(cmd string) (err error) {

		str, ok := gui.GUI.OutputReadString("Path to import")
		if !ok {
			return
		}

		data, err := os.ReadFile(str + ".pandora")
		if err != nil {
			return
		}

		if _, err = wallet.ImportWalletAddressJSON(data); err != nil {
			return
		}

		gui.GUI.Info("Imported successfully")
		return
	}

	cliExportWalletJSON := func(cmd string) (err error) {

		str, ok := gui.GUI.OutputReadString("Path to export")
		if !ok {
			return
		}

		f, err := os.Create(str + ".pandora")
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

		gui.GUI.Info("Wallet Exported successfully")
		return
	}

	cliImportWalletJSON := func(cmd string) (err error) {

		str, ok := gui.GUI.OutputReadString("Path to import Wallet")
		if !ok {
			return
		}

		done, ok := gui.GUI.OutputReadBool("Your wallet will be REPLACED with this one. Are you sure ?")
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

		if err = wallet.ImportWalletJSON(data); err != nil {
			return
		}

		gui.GUI.Info("Wallet Imported Successfully")
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
		if success, err = wallet.RemoveAddress(index, ""); err != nil {
			return
		}
		if err = wallet.CliListAddresses(""); err != nil {
			return
		}

		if success {
			gui.GUI.OutputWrite("Address removed")
		} else {
			gui.GUI.OutputWrite("Address was NOT removed ")
		}
		return
	}

	cliDeriveDelegatedStake := func(cmd string) (err error) {

		addr, _, err := wallet.CliSelectAddress("Select Address to Derive Delegated Stake")
		if err != nil {
			return
		}

		nonce, ok := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", nil, true)
		if !ok {
			return
		}

		return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

			chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

			accs := accounts.NewAccounts(reader)
			var acc *account.Account
			if acc, err = accs.GetAccount(addr.PublicKeyHash, chainHeight); err != nil {
				return
			}

			if nonce == 0 && acc != nil {
				nonce = wallet.mempool.GetNonce(addr.PublicKeyHash, acc.Nonce)
			}

			var delegatedStake *wallet_address.WalletAddressDelegatedStake
			if delegatedStake, err = addr.DeriveDelegatedStake(uint32(nonce)); err != nil {
				return
			}

			gui.GUI.OutputWrite("Delegated stake:")
			gui.GUI.OutputWrite("   PublicKeyHash", delegatedStake.PublicKeyHash)
			gui.GUI.OutputWrite("   PrivateKey", delegatedStake.PrivateKey.Key)

			str, ok := gui.GUI.OutputReadString("Path to export to a file")
			if !ok {
				return
			}

			if str != "" {
				var f *os.File
				if f, err = os.Create(str + ".delegatedStake"); err != nil {
					return
				}

				defer f.Close()

				delegatedStakeOut := struct {
					DelegatedStakePublicKeyHash helpers.HexBytes
					AddressPublicKeyHash        helpers.HexBytes
				}{
					delegatedStake.PublicKeyHash,
					addr.PublicKeyHash,
				}

				var marshal []byte
				if marshal, err = json.Marshal(delegatedStakeOut); err != nil {
					return
				}

				if _, err = fmt.Fprint(f, string(marshal)); err != nil {
					return errors.New("Error writing into file")
				}

			}

			return
		})

		return
	}

	cliShowMnemonic := func(string) (err error) {
		gui.GUI.OutputWrite("Mnemonic \n")
		gui.GUI.OutputWrite(wallet.Mnemonic)

		gui.GUI.OutputWrite("Seed \n")
		gui.GUI.OutputWrite(wallet.Seed)

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
		gui.GUI.OutputWrite(privateKey)

		return
	}

	cliImportPrivateKey := func(cmd string) (err error) {

		privateKey, ok := gui.GUI.OutputReadBytes("Write Private key", []int{32})
		if !ok {
			return
		}

		name, ok := gui.GUI.OutputReadString("Write Name of the newly imported address")
		if !ok {
			return
		}

		var adr *wallet_address.WalletAddress
		if adr, err = wallet.ImportPrivateKey(name, privateKey); err != nil {
			return
		}

		gui.GUI.OutputWrite("Address was imported: " + adr.AddressEncoded)

		return
	}

	cliEncryptWallet := func(cmd string) (err error) {
		password, ok := gui.GUI.OutputReadString("Write the password that will be used for encryption")
		if !ok {
			return
		}

		return wallet.Encryption.Encrypt(password)
	}

	cliDecyprtWallet := func(cmd string) (err error) {
		return
	}

	cliRemoveEncryption := func(cmd string) (err error) {
		return
	}

	gui.GUI.CommandDefineCallback("List Addresses", wallet.CliListAddresses)
	gui.GUI.CommandDefineCallback("Create New Address", cliCreateNewAddress)
	gui.GUI.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic)
	gui.GUI.CommandDefineCallback("Show Private Key", cliShowPrivateKey)
	gui.GUI.CommandDefineCallback("Import Private Key", cliImportPrivateKey)
	gui.GUI.CommandDefineCallback("Remove Address", cliRemoveAddress)
	gui.GUI.CommandDefineCallback("Derive Delegated Stake", cliDeriveDelegatedStake)
	gui.GUI.CommandDefineCallback("Export Address JSON", cliExportAddressJSON)
	gui.GUI.CommandDefineCallback("Import Address JSON", cliImportAddressJSON)
	gui.GUI.CommandDefineCallback("Export Wallet JSON", cliExportWalletJSON)
	gui.GUI.CommandDefineCallback("Import Wallet JSON", cliImportWalletJSON)
	gui.GUI.CommandDefineCallback("Encrypt Wallet", cliEncryptWallet)
	gui.GUI.CommandDefineCallback("Decrypt Wallet", cliDecyprtWallet)
	gui.GUI.CommandDefineCallback("Remove Encryption", cliRemoveEncryption)

}
