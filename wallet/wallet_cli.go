package wallet

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
	"strconv"
)

func (wallet *Wallet) deriveDelegatedStake(addr *wallet_address.WalletAddress, nonce uint64, path string, print bool) error {

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		plainAccs := plain_accounts.NewPlainAccounts(reader)

		var plainAcc *plain_account.PlainAccount
		if plainAcc, err = plainAccs.GetPlainAccount(addr.PublicKey, chainHeight); err != nil {
			return
		}

		if nonce == 0 && plainAcc != nil {
			nonce = wallet.mempool.GetNonce(addr.PublicKey, plainAcc.Nonce)
		}

		var delegatedStake *wallet_address.WalletAddressDelegatedStake
		if delegatedStake, err = addr.DeriveDelegatedStake(uint32(nonce)); err != nil {
			return
		}

		if print {
			gui.GUI.OutputWrite("Delegated stake:")
			gui.GUI.OutputWrite("   PublicKey", delegatedStake.PublicKey)
			gui.GUI.OutputWrite("   PrivateKey", delegatedStake.PrivateKey.Key)
		}

		if path != "" {

			var f *os.File
			if f, err = os.Create(path); err != nil {
				return
			}

			defer f.Close()

			delegatedStakeOut := &DelegatedStakeOutput{
				addr.AddressEncoded,
				delegatedStake.PublicKey,
			}

			var marshal []byte
			if marshal, err = json.Marshal(delegatedStakeOut); err != nil {
				return
			}

			if _, err = fmt.Fprint(f, string(marshal)); err != nil {
				return err
			}

		}

		return
	})
}

func (wallet *Wallet) CliListAddresses(cmd string, ctx context.Context) (err error) {

	wallet.Lock()
	defer wallet.Unlock()

	gui.GUI.OutputWrite("Wallet")
	gui.GUI.OutputWrite("Version: " + wallet.Version.String())
	gui.GUI.OutputWrite("Encrypted: " + wallet.Encryption.Encrypted.String())

	gui.GUI.OutputWrite("Count: " + strconv.Itoa(wallet.Count))
	gui.GUI.OutputWrite("")

	type AddressAsset struct {
		balance *crypto.ElGamal
		assetId []byte
		ast     *asset.Asset
	}
	type Address struct {
		assetsList []*AddressAsset
	}
	addresses := make([]*Address, len(wallet.Addresses))

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		dataStorage := data_storage.NewDataStorage(reader)

		var ast *asset.Asset
		var accs *accounts.Accounts
		var acc *account.Account

		for i, walletAddress := range wallet.Addresses {

			var isReg bool
			if isReg, err = dataStorage.Regs.Exists(string(walletAddress.PublicKey)); err != nil {
				return
			}

			addressStr := walletAddress.GetAddress(isReg)

			gui.GUI.OutputWrite(fmt.Sprintf("%d) %s : %s :: %s", i, walletAddress.Name, walletAddress.Version.String(), addressStr))

			switch walletAddress.Version {

			case wallet_address.VERSION_NORMAL:

				var assetsList [][]byte
				if assetsList, err = dataStorage.AccsCollection.GetAccountAssets(walletAddress.PublicKey); err != nil {
					return
				}

				var plainAcc *plain_account.PlainAccount
				if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(walletAddress.PublicKey, chainHeight); err != nil {
					return
				}

				if len(assetsList) == 0 && plainAcc == nil {
					gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
					continue
				}

				if plainAcc != nil {

					gui.GUI.OutputWrite(fmt.Sprintf("%18s: %d", "Nonce", plainAcc.Nonce))
					gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Unclaimed", strconv.FormatFloat(config_coins.ConvertToBase(plainAcc.Unclaimed), 'f', config_coins.DECIMAL_SEPARATOR, 64)))
					if plainAcc.DelegatedStake.HasDelegatedStake() {
						gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "Stake Available", strconv.FormatFloat(config_coins.ConvertToBase(plainAcc.DelegatedStake.StakeAvailable), 'f', config_coins.DECIMAL_SEPARATOR, 64)))

						if len(plainAcc.DelegatedStake.StakesPending) > 0 {
							gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES", ""))
							for _, stakePending := range plainAcc.DelegatedStake.StakesPending {
								gui.GUI.OutputWrite(fmt.Sprintf("%18s: %10s %t", strconv.FormatUint(stakePending.ActivationHeight, 10), strconv.FormatFloat(config_coins.ConvertToBase(stakePending.PendingAmount), 'f', config_coins.DECIMAL_SEPARATOR, 64), stakePending.PendingType))
							}
						} else {
							gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES:", "EMPTY"))
						}
					}

					if plainAcc.AssetFeeLiquidities.HasAssetFeeLiquidities() {

						gui.GUI.OutputWrite(fmt.Sprintf("%18s: %d", "Liquidities", len(plainAcc.AssetFeeLiquidities.List)))
						for i, assetFeeLiquidity := range plainAcc.AssetFeeLiquidities.List {
							gui.GUI.OutputWrite(fmt.Sprintf("%18s: %20s Rate %d LeadingZeros %d", strconv.Itoa(i), hex.EncodeToString(assetFeeLiquidity.Asset), assetFeeLiquidity.Rate, assetFeeLiquidity.LeadingZeros))
						}

					}

				}

				addresses[i] = &Address{}
				if len(assetsList) > 0 {

					gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s %d", "BALANCES ENCRYPTED", "", len(assetsList)))
					for _, assetId := range assetsList {

						if ast, err = dataStorage.Asts.GetAsset(assetId); err != nil {
							return
						}
						if accs, err = dataStorage.AccsCollection.GetMap(assetId); err != nil {
							return
						}

						if acc, err = accs.GetAccount(walletAddress.PublicKey); err != nil {
							return
						}
						gui.GUI.OutputWrite(fmt.Sprintf("%260s: %s", hex.EncodeToString(acc.Balance.Amount.Serialize()), ast.Name))

						addresses[i].assetsList = append(addresses[i].assetsList, &AddressAsset{
							acc.Balance.Amount,
							assetId,
							ast,
						})

					}

				}
			default:
				return errors.New("Invalid Address version")
			}

		}

		return
	}); err != nil {
		return
	}

	for i, walletAddress := range wallet.Addresses {
		for _, data := range addresses[i].assetsList {

			var decoded uint64
			if decoded, err = wallet.DecodeBalanceByPublicKey(walletAddress.PublicKey, data.balance, data.assetId, true, false, ctx, func(status string) {
				gui.GUI.Info2Update("Decoding", status)
			}); err != nil {
				return
			}

			gui.GUI.Info2Update("Decoding", "")

			gui.GUI.OutputWrite(fmt.Sprintf("%18s: %s", strconv.FormatFloat(config_coins.ConvertToBase(decoded), 'f', config_coins.DECIMAL_SEPARATOR, 64), data.ast.Name))
		}

		gui.GUI.OutputWrite(fmt.Sprintf("%18s", "DONE DECRYPTING"))

	}

	return
}

func (wallet *Wallet) CliSelectAddress(text string, ctx context.Context) (*wallet_address.WalletAddress, int, error) {

	if err := wallet.CliListAddresses("", ctx); err != nil {
		return nil, 0, err
	}

	index := gui.GUI.OutputReadInt(text, false, func(value int) bool {
		return value < wallet.GetAddressesCount()
	})

	walletAddress, err := wallet.GetWalletAddress(index)
	if err != nil {
		return nil, 0, err
	}

	return walletAddress, index, nil
}

func (wallet *Wallet) initWalletCLI() {

	cliExportAddresses := func(cmd string, ctx context.Context) (err error) {
		filename := gui.GUI.OutputReadFilename("Path to export", "txt")

		f, err := os.Create(filename)
		if err != nil {
			return
		}

		defer f.Close()

		wallet.RLock()
		defer wallet.RUnlock()

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

		index := gui.GUI.OutputReadInt("Select Address to be Exported", false, nil)
		filename := gui.GUI.OutputReadFilename("Path to export", "pandora")

		f, err := os.Create(filename)
		if err != nil {
			return
		}

		defer f.Close()

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

		wallet.RLock()
		defer wallet.RUnlock()

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

		done := gui.GUI.OutputReadBool("Your wallet will be REPLACED with this one! y/n")

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
		if _, err = wallet.AddNewAddress(true); err != nil {
			return
		}
		return wallet.CliListAddresses(cmd, ctx)
	}

	cliRemoveAddress := func(cmd string, ctx context.Context) (err error) {

		_, index, err := wallet.CliSelectAddress("Select Address to be Removed", ctx)
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

	cliDeriveDelegatedStake := func(cmd string, ctx context.Context) (err error) {

		addr, _, err := wallet.CliSelectAddress("Select Address to Derive Delegated Stake", ctx)
		if err != nil {
			return
		}

		nonce := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", true, nil)
		path := gui.GUI.OutputReadFilename("Path to export to a file", "delegatedStake")

		return wallet.deriveDelegatedStake(addr, nonce, path, true)

	}

	cliShowMnemonic := func(cmd string, ctx context.Context) (err error) {

		gui.GUI.OutputWrite("Mnemonic")
		gui.GUI.OutputWrite(wallet.Mnemonic)

		gui.GUI.OutputWrite("Seed")
		gui.GUI.OutputWrite(wallet.Seed)

		return
	}

	cliShowPrivateKey := func(cmd string, ctx context.Context) (err error) {

		_, index, err := wallet.CliSelectAddress("Select Address to show the private key", ctx)
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

	cliImportPrivateKey := func(cmd string, ctx context.Context) (err error) {

		privateKey := gui.GUI.OutputReadBytes("Write Private key", func(input []byte) bool {
			return len(input) == 32
		})

		name := gui.GUI.OutputReadString("Write Name of the newly imported address")

		var adr *wallet_address.WalletAddress
		if adr, err = wallet.ImportPrivateKey(name, privateKey); err != nil {
			return
		}

		gui.GUI.OutputWrite("Address was imported: " + adr.AddressEncoded)

		return
	}

	cliEncryptWallet := func(cmd string, ctx context.Context) (err error) {

		password := gui.GUI.OutputReadString("Password for encrypting wallet")
		difficulty := gui.GUI.OutputReadInt("Difficulty for encryption", false, func(value int) bool {
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
	gui.GUI.CommandDefineCallback("Show Private Key", cliShowPrivateKey, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Private Key", cliImportPrivateKey, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Remove Address", cliRemoveAddress, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Derive Delegated Stake", cliDeriveDelegatedStake, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Addresses", cliExportAddresses, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Address JSON", cliExportAddressJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Address JSON", cliImportAddressJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Export Wallet JSON", cliExportWalletJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Import Wallet JSON", cliImportWalletJSON, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Encrypt Wallet", cliEncryptWallet, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Remove Encryption", cliRemoveEncryption, wallet.Loaded)
	gui.GUI.CommandDefineCallback("Decrypt Wallet", cliDecryptWallet, !wallet.Loaded)

}
