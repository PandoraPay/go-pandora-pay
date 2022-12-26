package wallet

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"os"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/files"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
	"pandora-pay/wallet/wallet_address/shared_staked"
	"strconv"
)

func (wallet *Wallet) exportSharedStakedAddress(addr *wallet_address.WalletAddress, path string, print bool) (*shared_staked.WalletAddressSharedStakedAddressExported, error) {

	if !addr.Staked {
		return nil, errors.New("Address is not Staked")
	}

	if print {
		gui.GUI.OutputWrite("Address:")
		gui.GUI.OutputWrite("   Encoded", addr.AddressEncoded)
		gui.GUI.OutputWrite("   Encoded with Registration", addr.AddressRegistrationEncoded)
	}

	sharedStakedAddress := &shared_staked.WalletAddressSharedStakedAddressExported{addr.AddressRegistrationEncoded}

	if path != "" {

		bytes, err := json.Marshal(sharedStakedAddress)
		if err != nil {
			return nil, err
		}

		if err := files.WriteFile(path, string(bytes)); err != nil {
			return nil, err
		}

	}

	return sharedStakedAddress, nil
}

func (wallet *Wallet) CliScanAddresses(cmd string, ctx context.Context) (err error) {

	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		for {

			var addr *addresses.Address
			if addr, err = wallet.GenerateNextAddress(false); err != nil {
				return
			}

			var reg *registration.Registration
			if reg, err = dataStorage.Regs.Get(string(addr.PublicKey)); err != nil {
				return
			}

			if reg != nil {
				if _, err = wallet.AddNewAddress(false, "", reg.Staked, reg.SpendPublicKey != nil, true); err != nil {
					return
				}
				continue
			}

			var plainAcc *plain_account.PlainAccount
			if plainAcc, err = dataStorage.PlainAccs.Get(string(addr.PublicKey)); err != nil {
				return
			}

			if plainAcc != nil {
				if _, err = wallet.AddNewAddress(false, "", false, false, true); err != nil {
					return
				}
				continue
			}

			return
		}

	}); err != nil {
		return
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
		registration            *registration.Registration
		plainAcc                *plain_account.PlainAccount
		assetsList              []*AddressAsset
		publicKey               []byte
		name                    string
		addressString           string
		addressRegisteredString string
	}

	wallet.Lock.RLock()
	gui.GUI.OutputWrite("Wallet")
	gui.GUI.OutputWrite("Version: " + wallet.Version.String())
	gui.GUI.OutputWrite("Encrypted: " + wallet.Encryption.Encrypted.String())

	gui.GUI.OutputWrite("Count: " + strconv.Itoa(wallet.Count))
	gui.GUI.OutputWrite("")

	addresses := make([]*Address, len(wallet.Addresses))

	for i, walletAddress := range wallet.Addresses {
		addresses[i] = &Address{publicKey: helpers.CloneBytes(walletAddress.PublicKey), name: walletAddress.Name, addressString: walletAddress.GetAddress(false), addressRegisteredString: walletAddress.GetAddress(true)}
	}
	wallet.Lock.RUnlock()

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		var ast *asset.Asset
		var accs *accounts.Accounts
		var acc *account.Account

		for i, address := range addresses {

			if addresses[i].registration, err = dataStorage.Regs.Get(string(address.publicKey)); err != nil {
				return
			}

			var assetsList [][]byte
			if assetsList, err = dataStorage.AccsCollection.GetAccountAssets(address.publicKey); err != nil {
				return
			}

			if addresses[i].plainAcc, err = dataStorage.PlainAccs.Get(string(address.publicKey)); err != nil {
				return
			}

			if len(assetsList) > 0 {

				for _, assetId := range assetsList {

					if ast, err = dataStorage.Asts.Get(string(assetId)); err != nil {
						return
					}
					if accs, err = dataStorage.AccsCollection.GetMap(assetId); err != nil {
						return
					}

					if acc, err = accs.Get(string(address.publicKey)); err != nil {
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

		if addresses[i].registration != nil {
			gui.GUI.OutputWrite(fmt.Sprintf("%d) %s :: %s", i, address.name, address.addressRegisteredString))
		} else {
			gui.GUI.OutputWrite(fmt.Sprintf("%d) %s :: %s", i, address.name, address.addressString))
		}

		if len(addresses[i].assetsList) == 0 && addresses[i].plainAcc == nil {
			gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
			continue
		}

		if addresses[i].registration != nil {
			gui.GUI.OutputWrite(fmt.Sprintf("%18s: Staked: %v SpendPublicKey: %s", "Registered", addresses[i].registration.Staked, base64.StdEncoding.EncodeToString(addresses[i].registration.SpendPublicKey)))
		}

		if addresses[i].plainAcc != nil {

			gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Unclaimed", strconv.FormatFloat(config_coins.ConvertToBase(addresses[i].plainAcc.Unclaimed), 'f', config_coins.DECIMAL_SEPARATOR, 64)))

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
		filename := gui.GUI.OutputReadFilename("Path to export", "txt", false)

		lines := []string{}

		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()

		if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
			regs := registrations.NewRegistrations(reader)

			for _, walletAddress := range wallet.Addresses {

				var isReg bool
				if isReg, err = regs.Exists(string(walletAddress.PublicKey)); err != nil {
					return
				}

				addressStr := walletAddress.GetAddress(isReg) + config.LineBreak
				lines = append(lines, addressStr)
			}

			return
		}); err != nil {
			return
		}

		if err := files.WriteFile(filename, lines...); err != nil {
			return err
		}

		gui.GUI.OutputWrite("Exported successfully to: ", filename)
		return
	}

	cliExportAddressJSON := func(cmd string, ctx context.Context) (err error) {

		if err = wallet.CliListAddresses("", ctx); err != nil {
			return
		}

		index := gui.GUI.OutputReadInt("Select Address to be Exported", false, 0, nil)
		filename := gui.GUI.OutputReadFilename("Path to export", "pandora", false)

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

		if err = files.WriteFile(filename, string(marshal)); err != nil {
			return err
		}

		gui.GUI.OutputWrite("Exported successfully to: ", filename)
		return
	}

	cliImportAddressJSON := func(cmd string, ctx context.Context) (err error) {

		str := gui.GUI.OutputReadFilename("Path to import Address", "pandora", false)

		data, err := os.ReadFile(str)
		if err != nil {
			return
		}

		if _, err = wallet.ImportWalletAddressJSON(data); err != nil {
			return
		}

		gui.GUI.OutputWrite("Imported successfully from: ", str)
		return
	}

	cliExportWalletJSON := func(cmd string, ctx context.Context) (err error) {

		filename := gui.GUI.OutputReadFilename("Path to export", "pandorawallet", false)

		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()

		var marshal []byte
		if marshal, err = json.Marshal(wallet); err != nil {
			return errors.New("Error marshaling wallet")
		}

		if err = files.WriteFile(filename, string(marshal)); err != nil {
			return
		}

		gui.GUI.OutputWrite("Wallet Exported successfully to: ", filename)
		return
	}

	cliImportWalletJSON := func(cmd string, ctx context.Context) (err error) {

		str := gui.GUI.OutputReadFilename("Path to import Wallet", "pandorawallet", false)

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

		gui.GUI.OutputWrite("Wallet Imported Successfully from: ", str)
		return
	}

	cliCreateNewAddress := func(cmd string, ctx context.Context) (err error) {

		filename := gui.GUI.OutputReadFilename("Name of your new address", "", false)
		staked := gui.GUI.OutputReadBool("Staked address ? y/n. Leave empty for n", true, false)
		spendRequired := gui.GUI.OutputReadBool("Spend Key required ? y/n. Leave empty for n", true, false)

		if _, err = wallet.AddNewAddress(true, filename, staked, spendRequired, true); err != nil {
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

		filename := gui.GUI.OutputReadFilename("Path to export to a file", "staked", false)

		_, err = wallet.exportSharedStakedAddress(addr, filename, true)
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

		gui.GUI.OutputWrite("WARNING!!! THIS COMMAND WILL DELETE YOUR EXISTING WALLET!", config.LineBreak, config.LineBreak)

		if !gui.GUI.OutputReadBool("Are you sure you want to clear the existing wallet and get a new one? y/n", false, false) {
			return
		}

		if err = wallet.CreateEmptyWallet(); err != nil {
			return
		}

		gui.GUI.OutputWrite("A new wallet has been created!")

		return
	}

	cliImportMnemonic := func(cmd string, ctx context.Context) (err error) {
		gui.GUI.OutputWrite("WARNING!!! THIS COMMAND WILL DELETE YOUR EXISTING WALLET!", config.LineBreak, config.LineBreak)

		if !gui.GUI.OutputReadBool("Are you sure you want to clear the existing wallet and import a mnemonic? y/n", false, false) {
			return
		}

		mnemonic := gui.GUI.OutputReadString("Provide the mnemonic")

		if err = wallet.ImportMnemonic(mnemonic); err != nil {
			return
		}

		gui.GUI.OutputWrite("A new wallet has been created using the mnemonic provided!")

		return
	}

	cliImportEntropy := func(cmd string, ctx context.Context) (err error) {

		gui.GUI.OutputWrite("WARNING!!! THIS COMMAND WILL DELETE YOUR EXISTING WALLET!", config.LineBreak, config.LineBreak)

		if !gui.GUI.OutputReadBool("Are you sure you want to clear the existing wallet and import an entropy? y/n", false, false) {
			return
		}

		entropy := gui.GUI.OutputReadBytes("Provide the entropy", func(b []byte) bool {
			return len(b) == 16 || len(b) == 32
		})

		if err = wallet.ImportEntropy(entropy); err != nil {
			return
		}

		gui.GUI.OutputWrite("A new wallet has been created using the seed provided!")

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
		staked := gui.GUI.OutputReadBool("Staked address ? y/n. Leave empty for n", true, false)
		spendRequired := gui.GUI.OutputReadBool("Spend Key required ? y/n. Leave empty for n", true, false)

		var adr *wallet_address.WalletAddress
		if adr, err = wallet.ImportSecretKey(name, secretKey, staked, spendRequired); err != nil {
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

	cliCreatePair := func(cmd string, ctx context.Context) (err error) {
		key := addresses.GenerateNewPrivateKey()
		pub := key.GeneratePublicKey()

		gui.GUI.OutputWrite("PRIVATE KEY", key.Key)
		gui.GUI.OutputWrite("PUBLIC KEY", pub, config.LineBreak, config.LineBreak)

		if filename := gui.GUI.OutputReadFilename("Path to export", "txt", true); len(filename) > 0 {

			if err = files.WriteFile(filename, fmt.Sprintf("PRIVATE KEY: %s %s", base64.StdEncoding.EncodeToString(key.Key), config.LineBreak), fmt.Sprintf("PUBLIC KEY: %s %s", base64.StdEncoding.EncodeToString(pub), config.LineBreak)); err != nil {
				return
			}

			gui.GUI.OutputWrite("Exported successfully to: ", filename)
		}
		return
	}

	cliSignMessage := func(cmd string, ctx context.Context) (err error) {
		privatekey := gui.GUI.OutputReadBytes("Private Key", func(val []byte) bool {
			return len(val) == cryptography.PrivateKeySize
		})

		var message []byte

		if gui.GUI.OutputReadBool("Base64 message y/n. Leave empty for yes.", true, true) {
			message = gui.GUI.OutputReadBytes("Message", nil)
		} else {
			message = []byte(gui.GUI.OutputReadString("Message"))
		}

		if gui.GUI.OutputReadBool("Hashing message using SHA3 y/n. Leave empty for yes.", true, true) {
			message = cryptography.SHA3(message)
		}

		signature, err := crypto.SignMessage(message, privatekey)
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Signature: ", signature, config.LineBreak, config.LineBreak)

		if filename := gui.GUI.OutputReadFilename("Path to export", "txt", true); len(filename) > 0 {

			if err = files.WriteFile(filename, fmt.Sprintf("Signature: %s %s", base64.StdEncoding.EncodeToString(signature), config.LineBreak)); err != nil {
				return
			}

			gui.GUI.OutputWrite("Exported successfully to: ", filename)
		}

		return
	}

	cliSignResolutionConditionalPayment := func(cmd string, ctx context.Context) (err error) {

		extra := &transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment{}

		privateKey := gui.GUI.OutputReadBytes("Private Key", func(value []byte) bool {
			return len(value) == cryptography.PrivateKeySize
		})
		pk, err := addresses.NewPrivateKey(privateKey)
		if err != nil {
			return
		}

		extra.TxId = gui.GUI.OutputReadBytes("Provide TxId", func(val []byte) bool {
			return len(val) == cryptography.HashSize
		})

		extra.PayloadIndex = byte(gui.GUI.OutputReadInt("Payload index", false, 0, func(val int) bool {
			return val >= 0 && val < 255
		}))

		extra.Resolution = gui.GUI.OutputReadBool("Resolution.  Use y/n for voting", false, false)

		signature, err := crypto.SignMessage(extra.MessageForSigning(), privateKey)
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Public Key: %s", base64.StdEncoding.EncodeToString(pk.GeneratePublicKey())))
		gui.GUI.OutputWrite(fmt.Sprintf("Signature: %s", base64.StdEncoding.EncodeToString(signature)))

		if filename := gui.GUI.OutputReadFilename("Path to export", "txt", true); len(filename) > 0 {

			if err = files.WriteFile(filename, fmt.Sprintf("Public Key: %s %s", base64.StdEncoding.EncodeToString(pk.GeneratePublicKey()), config.LineBreak),
				fmt.Sprintf("Signature: %s %s", base64.StdEncoding.EncodeToString(signature), config.LineBreak), config.LineBreak); err != nil {
				return
			}

			gui.GUI.OutputWrite("Exported successfully to: ", filename)
		}

		return
	}

	gui.GUI.CommandDefineCallback("List Addresses", wallet.CliListAddresses, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Scan Addresses", wallet.CliScanAddresses, wallet.Loaded)
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

	gui.GUI.CommandDefineCallback("Create (PublicKey, PrivateKey) pair", cliCreatePair, true)
	gui.GUI.CommandDefineCallback("Sign message using PrivateKey", cliSignMessage, true)
	gui.GUI.CommandDefineCallback("Sign Resolution Conditional Payment", cliSignResolutionConditionalPayment, true)

}
