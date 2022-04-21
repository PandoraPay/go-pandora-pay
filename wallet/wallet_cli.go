package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"os"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_coins"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
	"pandora-pay/wallet/wallet_address/shared_staked"
	"strconv"
)

func (wallet *Wallet) exportSharedStakedAddress(addr *wallet_address.WalletAddress, path string, print bool) (*shared_staked.WalletAddressSharedStakedAddressExported, error) {

	if print {
		gui.GUI.OutputWrite("Address:")
		gui.GUI.OutputWrite("   Encoded", addr.AddressEncoded)
	}

	sharedStaked, err := addr.DeriveSharedStaked(0)
	if err != nil {
		return nil, err
	}

	sharedStakedAddress := &shared_staked.WalletAddressSharedStakedAddressExported{addr.AddressEncoded, sharedStaked.PublicKey}

	if path != "" {

		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		defer f.Close()

		bytes, err := json.Marshal(sharedStakedAddress)
		if err != nil {
			return nil, err
		}

		if _, err = fmt.Fprint(f, string(bytes)); err != nil {
			return nil, err
		}
	}

	return sharedStakedAddress, nil
}

func (wallet *Wallet) CliListAddresses(cmd string, ctx context.Context) (err error) {

	type AddressAsset struct {
		balance uint64
		assetId []byte
		ast     *asset.Asset
	}
	type Address struct {
		plainAcc      *plain_account.PlainAccount
		assetsList    []*AddressAsset
		publicKeyHash []byte
		name          string
		addressString string
	}

	wallet.Lock.RLock()
	gui.GUI.OutputWrite("Wallet")
	gui.GUI.OutputWrite("Version: " + wallet.Version.String())
	gui.GUI.OutputWrite("Encrypted: " + wallet.Encryption.Encrypted.String())

	gui.GUI.OutputWrite("Count: " + strconv.Itoa(wallet.Count))
	gui.GUI.OutputWrite("")

	addresses := make([]*Address, len(wallet.Addresses))

	for i, walletAddress := range wallet.Addresses {
		addresses[i] = &Address{publicKeyHash: helpers.CloneBytes(walletAddress.PublicKeyHash), name: walletAddress.Name, addressString: walletAddress.GetAddress()}
	}
	wallet.Lock.RUnlock()

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		var ast *asset.Asset
		var accs *accounts.Accounts
		var acc *account.Account

		for i, address := range addresses {

			var assetsList [][]byte
			if assetsList, err = dataStorage.AccsCollection.GetAccountAssets(address.publicKeyHash); err != nil {
				return
			}

			if addresses[i].plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(address.publicKeyHash); err != nil {
				return
			}

			if len(assetsList) > 0 {

				for _, assetId := range assetsList {

					if ast, err = dataStorage.Asts.GetAsset(assetId); err != nil {
						return
					}
					if accs, err = dataStorage.AccsCollection.GetMap(assetId); err != nil {
						return
					}

					if acc, err = accs.GetAccount(address.publicKeyHash); err != nil {
						return
					}

					addresses[i].assetsList = append(addresses[i].assetsList, &AddressAsset{
						acc.Balance,
						assetId,
						ast,
					})

				}

			}

		}

		return
	}); err != nil {
		return
	}

	for i, address := range addresses {

		gui.GUI.OutputWrite(fmt.Sprintf("%d) %s :: %s", i, address.name, address.addressString))

		if len(addresses[i].assetsList) == 0 && addresses[i].plainAcc == nil {
			gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
			continue
		}

		if addresses[i].plainAcc != nil {
			gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Staked", strconv.FormatFloat(config_coins.ConvertToBase(addresses[i].plainAcc.StakeAvailable), 'f', config_coins.DECIMAL_SEPARATOR, 64)))
		}

		if len(addresses[i].assetsList) > 0 {
			for _, data := range addresses[i].assetsList {
				gui.GUI.OutputWrite(fmt.Sprintf("%18s: %18s", data.ast.Name, strconv.FormatFloat(config_coins.ConvertToBase(data.balance), 'f', config_coins.DECIMAL_SEPARATOR, 64)))
			}
		}

		gui.GUI.Info2Update("Decoding", "")

	}

	gui.GUI.OutputWrite(fmt.Sprintf("%18s", "DONE"))

	return
}

func (wallet *Wallet) CliSelectAddress(text string, ctx context.Context) (*wallet_address.WalletAddress, string, int, error) {

	if err := wallet.CliListAddresses("", ctx); err != nil {
		return nil, "", 0, err
	}

	index := gui.GUI.OutputReadInt(text, false, 0, func(value int) bool {
		return value < wallet.GetAddressesCount()
	})

	walletAddress, err := wallet.GetWalletAddress(index, true)
	if err != nil {
		return nil, "", 0, err
	}

	return walletAddress, walletAddress.AddressEncoded, index, nil
}

func (wallet *Wallet) initWalletCLI() {

	cliExportAddresses := func(cmd string, ctx context.Context) (err error) {
		filename := gui.GUI.OutputReadFilename("Path to export", "txt")

		f, err := os.Create(filename)
		if err != nil {
			return
		}

		defer f.Close()

		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()

		if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

			for _, walletAddress := range wallet.Addresses {

				addressStr := walletAddress.GetAddress()
				if _, err = fmt.Fprintln(f, addressStr); err != nil {
					return
				}
			}

			return
		}); err != nil {
			return
		}

		gui.GUI.Info("Exported successfully to: ", filename)
		return
	}

	cliExportAddressJSON := func(cmd string, ctx context.Context) (err error) {

		if err = wallet.CliListAddresses("", ctx); err != nil {
			return
		}

		index := gui.GUI.OutputReadInt("Select Address to be Exported", false, 0, nil)
		filename := gui.GUI.OutputReadFilename("Path to export", "pandora")

		f, err := os.Create(filename)
		if err != nil {
			return
		}

		defer f.Close()

		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()

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
		if marshal, err = wallet.Encryption.encryptData(marshal); err != nil {
			return
		}

		if _, err = fmt.Fprint(f, string(marshal)); err != nil {
			return
		}

		gui.GUI.Info("Exported successfully to: ", filename)
		return
	}

	cliImportAddressJSON := func(cmd string, ctx context.Context) (err error) {

		str := gui.GUI.OutputReadFilename("Path to import Address", "pandora")

		data, err := os.ReadFile(str)
		if err != nil {
			return
		}

		if _, err = wallet.ImportWalletAddressJSON(data); err != nil {
			return
		}

		gui.GUI.Info("Imported successfully from: ", str)
		return
	}

	cliExportWalletJSON := func(cmd string, ctx context.Context) (err error) {

		filename := gui.GUI.OutputReadFilename("Path to export", "pandorawallet")

		f, err := os.Create(filename)
		if err != nil {
			return
		}

		defer f.Close()

		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()

		var marshal []byte
		if marshal, err = json.Marshal(wallet); err != nil {
			return errors.New("Error marshaling wallet")
		}

		if _, err = fmt.Fprint(f, string(marshal)); err != nil {
			return
		}

		gui.GUI.Info("Wallet Exported successfully to: ", filename)
		return
	}

	cliImportWalletJSON := func(cmd string, ctx context.Context) (err error) {

		str := gui.GUI.OutputReadFilename("Path to import Wallet", "pandorawallet")

		done := gui.GUI.OutputReadBool("Your wallet will be REPLACED with this one! y/n", false, false)

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

		gui.GUI.Info("Wallet Imported Successfully from: ", str)
		return
	}

	cliCreateNewAddress := func(cmd string, ctx context.Context) (err error) {

		name := gui.GUI.OutputReadFilename("Name of your new address", "")

		if _, err = wallet.AddNewAddress(true, name, true); err != nil {
			return
		}
		return wallet.CliListAddresses(cmd, ctx)
	}

	cliRemoveAddress := func(cmd string, ctx context.Context) (err error) {

		_, _, index, err := wallet.CliSelectAddress("Select Address to be Removed", ctx)
		if err != nil {
			return
		}

		var success bool
		if success, err = wallet.RemoveAddressByIndex(index, true); err != nil {
			return
		}
		if err = wallet.CliListAddresses("", ctx); err != nil {
			return
		}

		if success {
			gui.GUI.OutputWrite("Address removed")
		} else {
			gui.GUI.OutputWrite("Address was NOT removed ")
		}
		return
	}

	cliExportSharedStakedAddress := func(cmd string, ctx context.Context) (err error) {

		addr, _, _, err := wallet.CliSelectAddress("Select Address to Export Shared Staked Address", ctx)
		if err != nil {
			return
		}

		path := gui.GUI.OutputReadFilename("Path to export to a file", "staked")

		_, err = wallet.exportSharedStakedAddress(addr, path, true)
		return err

	}

	cliShowMnemonic := func(cmd string, ctx context.Context) (err error) {

		gui.GUI.OutputWrite("Mnemonic")
		gui.GUI.OutputWrite("---------------------")
		gui.GUI.OutputWrite(wallet.Mnemonic)

		return
	}

	cliShowEntropy := func(cmd string, ctx context.Context) (err error) {

		gui.GUI.OutputWrite("Entropy")
		gui.GUI.OutputWrite("---------------------")
		entropy, err := bip39.EntropyFromMnemonic(wallet.Mnemonic)
		if err != nil {
			return
		}
		gui.GUI.OutputWrite(entropy)

		return
	}

	cliClearWallet := func(cmd string, ctx context.Context) (err error) {

		gui.GUI.OutputWrite("WARNING!!! THIS COMMAND WILL DELETE YOUR EXISTING WALLET!\n\n")

		if !gui.GUI.OutputReadBool("Are you sure you want to clear the existing wallet and get a new one? y/n", false, false) {
			return
		}

		if err = wallet.CreateEmptyWallet(); err != nil {
			return
		}

		gui.GUI.Info("A new wallet has been created!")

		return
	}

	cliImportMnemonic := func(cmd string, ctx context.Context) (err error) {
		gui.GUI.OutputWrite("WARNING!!! THIS COMMAND WILL DELETE YOUR EXISTING WALLET!\n\n")

		if !gui.GUI.OutputReadBool("Are you sure you want to clear the existing wallet and import a mnemonic? y/n", false, false) {
			return
		}

		mnemonic := gui.GUI.OutputReadString("Provide the mnemonic")

		if err = wallet.ImportMnemonic(mnemonic); err != nil {
			return
		}

		gui.GUI.Info("A new wallet has been created using the mnemonic provided!")

		return
	}

	cliImportEntropy := func(cmd string, ctx context.Context) (err error) {

		gui.GUI.OutputWrite("WARNING!!! THIS COMMAND WILL DELETE YOUR EXISTING WALLET!\n\n")

		if !gui.GUI.OutputReadBool("Are you sure you want to clear the existing wallet and import an entropy? y/n", false, false) {
			return
		}

		entropy := gui.GUI.OutputReadBytes("Provide the entropy", func(b []byte) bool {
			return len(b) == 16 || len(b) == 32
		})

		if err = wallet.ImportEntropy(entropy); err != nil {
			return
		}

		gui.GUI.Info("A new wallet has been created using the seed provided!")

		return
	}

	cliShowAddressSecretKey := func(cmd string, ctx context.Context) (err error) {

		_, _, index, err := wallet.CliSelectAddress("Select Address to show the secret key", ctx)
		if err != nil {
			return
		}

		secret, err := wallet.GetAddressSecretKey(index)
		if err != nil {
			return
		}
		gui.GUI.OutputWrite(secret)

		return
	}

	cliImportAddressSecretKey := func(cmd string, ctx context.Context) (err error) {

		secretKey := gui.GUI.OutputReadBytes("Write Secret key", func(input []byte) bool {
			return len(input) > 80
		})

		name := gui.GUI.OutputReadString("Write Name of the newly imported address")

		var adr *wallet_address.WalletAddress
		if adr, err = wallet.ImportSecretKey(name, secretKey); err != nil {
			return
		}

		gui.GUI.OutputWrite("Address was imported: " + adr.AddressEncoded)

		return
	}

	cliEncryptWallet := func(cmd string, ctx context.Context) (err error) {

		password := gui.GUI.OutputReadString("Password for encrypting wallet")
		difficulty := gui.GUI.OutputReadInt("Difficulty for encryption", false, 0, func(value int) bool {
			return value >= 1 && value <= 10
		})

		gui.GUI.OutputWrite("Wallet encrypting...")

		if err = wallet.Encryption.Encrypt(password, difficulty); err == nil {
			gui.GUI.OutputWrite("Wallet encrypted successfully")
		}
		return
	}

	cliDecryptWallet := func(cmd string, ctx context.Context) (err error) {

		password := gui.GUI.OutputReadString("Password for decrypting wallet")

		gui.GUI.OutputWrite("Wallet decrypting...")

		if err = wallet.Encryption.Decrypt(password); err == nil {
			gui.GUI.OutputWrite("Wallet decrypted successfully")
		}
		return
	}

	cliRemoveEncryption := func(cmd string, ctx context.Context) (err error) {
		gui.GUI.OutputWrite("Wallet removing encryption...")
		if err = wallet.Encryption.RemoveEncryption(); err == nil {
			gui.GUI.OutputWrite("Wallet encryption was removed successfully")
		}
		return
	}

	gui.GUI.CommandDefineCallback("List Addresses", wallet.CliListAddresses, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Create New Address", cliCreateNewAddress, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Clear & Create new empty Wallet", cliClearWallet, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Mnemnonic", cliImportMnemonic, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Show Entropy", cliShowEntropy, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Entropy", cliImportEntropy, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Show Address Secret Key", cliShowAddressSecretKey, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Address Secret Key", cliImportAddressSecretKey, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Remove Address", cliRemoveAddress, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Staked Staked Address", cliExportSharedStakedAddress, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Addresses", cliExportAddresses, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Address JSON", cliExportAddressJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Address JSON", cliImportAddressJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Wallet JSON", cliExportWalletJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Wallet JSON", cliImportWalletJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Encrypt Wallet", cliEncryptWallet, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Remove Encryption", cliRemoveEncryption, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Decrypt Wallet", cliDecryptWallet, !wallet.Loaded)

}
