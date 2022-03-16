package wallet

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
	"strconv"
)

func (wallet *Wallet) exportDelegatedAddress(addr *wallet_address.WalletAddress, path string, print bool) (err error) {

	if !addr.Stakable {
		return errors.New("Address is not VERSION_DELEGATED_STAKE")
	}

	if print {
		gui.GUI.OutputWrite("Address:")
		gui.GUI.OutputWrite("   Encoded", addr.AddressEncoded)
		gui.GUI.OutputWrite("   Encoded with Registration", addr.AddressRegistrationEncoded)
	}

	if path != "" {

		var f *os.File
		if f, err = os.Create(path); err != nil {
			return
		}

		defer f.Close()

		if _, err = fmt.Fprint(f, addr.AddressRegistrationEncoded); err != nil {
			return err
		}
	}

	return
}

func (wallet *Wallet) CliListAddresses(cmd string, ctx context.Context) (err error) {

	type AddressAsset struct {
		balance *crypto.ElGamal
		assetId []byte
		ast     *asset.Asset
	}
	type Address struct {
		isReg         bool
		plainAcc      *plain_account.PlainAccount
		assetsList    []*AddressAsset
		publicKey     []byte
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
		addresses[i] = &Address{publicKey: helpers.CloneBytes(walletAddress.PublicKey), name: walletAddress.Name, addressString: walletAddress.GetAddress(false)}
	}
	wallet.Lock.RUnlock()

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		dataStorage := data_storage.NewDataStorage(reader)

		var ast *asset.Asset
		var accs *accounts.Accounts
		var acc *account.Account

		for i, address := range addresses {

			if addresses[i].isReg, err = dataStorage.Regs.Exists(string(address.publicKey)); err != nil {
				return
			}

			var assetsList [][]byte
			if assetsList, err = dataStorage.AccsCollection.GetAccountAssets(address.publicKey); err != nil {
				return
			}

			if addresses[i].plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(address.publicKey, chainHeight); err != nil {
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

					if acc, err = accs.GetAccount(address.publicKey); err != nil {
						return
					}

					addresses[i].assetsList = append(addresses[i].assetsList, &AddressAsset{
						acc.Balance.Amount,
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

	var decrypted uint64
	for i, address := range addresses {

		gui.GUI.OutputWrite(fmt.Sprintf("%d) %s :: %s", i, address.name, address.addressString))

		if len(addresses[i].assetsList) == 0 && addresses[i].plainAcc == nil {
			gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
			continue
		}

		if addresses[i].plainAcc != nil {

			//gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Unclaimed", strconv.FormatFloat(config_coins.ConvertToBase(addresses[i].plainAcc.Unclaimed), 'f', config_coins.DECIMAL_SEPARATOR, 64)))

			if addresses[i].plainAcc.AssetFeeLiquidities.HasAssetFeeLiquidities() {

				gui.GUI.OutputWrite(fmt.Sprintf("%18s: %d", "Liquidities", len(addresses[i].plainAcc.AssetFeeLiquidities.List)))
				for i, assetFeeLiquidity := range addresses[i].plainAcc.AssetFeeLiquidities.List {
					gui.GUI.OutputWrite(fmt.Sprintf("%18s: %20s Rate %d LeadingZeros %d", strconv.Itoa(i), base64.StdEncoding.EncodeToString(assetFeeLiquidity.Asset), assetFeeLiquidity.Rate, assetFeeLiquidity.LeadingZeros))
				}

			}

		}

		if len(addresses[i].assetsList) > 0 {

			gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s %d", "BALANCES ENCRYPTED", "", len(addresses[i].assetsList)))
			for _, data := range addresses[i].assetsList {
				gui.GUI.OutputWrite(fmt.Sprintf("%18s: %64s", data.ast.Name, base64.StdEncoding.EncodeToString(data.balance.Serialize())))
			}

			gui.GUI.OutputWrite(fmt.Sprintf("%18s", "Decrypting...."))

			for _, data := range addresses[i].assetsList {
				gui.GUI.Info2Update("Decrypting", "")

				if decrypted, err = wallet.DecryptBalanceByPublicKey(address.publicKey, data.balance.Serialize(), data.assetId, false, 0, true, true, ctx, func(status string) {
					gui.GUI.Info2Update("Decrypted", status)
				}); err != nil {
					return
				}

				gui.GUI.OutputWrite(fmt.Sprintf("%18s: %18s", data.ast.Name, strconv.FormatFloat(config_coins.ConvertToBase(decrypted), 'f', config_coins.DECIMAL_SEPARATOR, 64)))
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
			regs := registrations.NewRegistrations(reader)

			for _, walletAddress := range wallet.Addresses {

				var isReg bool
				if isReg, err = regs.Exists(string(walletAddress.PublicKey)); err != nil {
					return
				}

				addressStr := walletAddress.GetAddress(isReg)
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

		if _, err = wallet.AddNewAddress(true, name, false, false); err != nil {
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

	cliExportDelegatedAddress := func(cmd string, ctx context.Context) (err error) {

		addr, _, _, err := wallet.CliSelectAddress("Select Address to Export Delegated Address", ctx)
		if err != nil {
			return
		}

		path := gui.GUI.OutputReadFilename("Path to export to a file", "delegatedStake")

		return wallet.exportDelegatedAddress(addr, path, true)

	}

	cliShowMnemonic := func(cmd string, ctx context.Context) (err error) {

		gui.GUI.OutputWrite("Mnemonic")
		gui.GUI.OutputWrite("---------------------")
		gui.GUI.OutputWrite(wallet.Mnemonic)

		gui.GUI.OutputWrite("\n")

		gui.GUI.OutputWrite("Seed")
		gui.GUI.OutputWrite("---------------------")
		gui.GUI.OutputWrite(wallet.Seed)

		return
	}

	cliShowSecretKey := func(cmd string, ctx context.Context) (err error) {

		_, _, index, err := wallet.CliSelectAddress("Select Address to show the secret key", ctx)
		if err != nil {
			return
		}

		privateKey, err := wallet.GetSecretKey(index)
		if err != nil {
			return
		}
		gui.GUI.OutputWrite(privateKey)

		return
	}

	cliImportSecretKey := func(cmd string, ctx context.Context) (err error) {

		secretKey := gui.GUI.OutputReadBytes("Write Secret key", func(input []byte) bool {
			return len(input) > 80
		})

		name := gui.GUI.OutputReadString("Write Name of the newly imported address")

		var adr *wallet_address.WalletAddress
		if adr, err = wallet.ImportSecretKey(name, secretKey, false, false); err != nil {
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
	gui.GUI.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Show Secret Key", cliShowSecretKey, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Secret Key", cliImportSecretKey, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Remove Address", cliRemoveAddress, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Delegated Address", cliExportDelegatedAddress, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Addresses", cliExportAddresses, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Address JSON", cliExportAddressJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Address JSON", cliImportAddressJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Wallet JSON", cliExportWalletJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Wallet JSON", cliImportWalletJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Encrypt Wallet", cliEncryptWallet, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Remove Encryption", cliRemoveEncryption, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Decrypt Wallet", cliDecryptWallet, !wallet.Loaded)

}
