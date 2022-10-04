package txs_builder

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
	"pandora-pay/config/config_assets"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/files"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/txs_builder/wizard"
)

func (builder *TxsBuilder) showWarningIfNotSyncCLI() {

}

func (builder *TxsBuilder) readData() (out *wizard.WizardTransactionData) {

	data := &wizard.WizardTransactionData{}
	str := gui.GUI.OutputReadString("Message (data). Leave empty for none")

	if len(str) > 0 {
		data.Data = []byte(str)
		data.Encrypt = gui.GUI.OutputReadBool("Encrypt message (data)? y/n. Leave empty for no", true, false)
	}

	return data
}

func (builder *TxsBuilder) readAmount(assetId []byte, text string) (amount uint64, err error) {

	amountFloat := gui.GUI.OutputReadFloat64(text, false, 0, nil)

	err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		asts := assets.NewAssets(reader)
		var ast *asset.Asset
		if ast, err = asts.Get(string(assetId)); err != nil {
			return
		}
		if ast == nil {
			return errors.New("Asset was not found")
		}
		if amount, err = ast.ConvertToUnits(amountFloat); err != nil {
			return
		}

		return
	})

	return
}

func (builder *TxsBuilder) readAddress(text string, leaveEmpty bool) (address *addresses.Address, err error) {

	for {
		str := gui.GUI.OutputReadString(text)
		if leaveEmpty && len(str) == 0 {
			break
		}

		address, err = addresses.DecodeAddr(str)
		if err != nil {
			gui.GUI.OutputWrite("Invalid Address")
			continue
		}
		break
	}

	return
}

func (builder *TxsBuilder) readAddressOptional(text string, assetId []byte, allowRandomAddress bool) (address *addresses.Address, addressEncoded string, amount uint64, err error) {

	text2 := text
	if allowRandomAddress {
		text2 = text + ". Leave empty for none"
	}

	for {
		str := gui.GUI.OutputReadString(text2)
		if str == "" && allowRandomAddress {
			return
		}

		if address, err = addresses.DecodeAddr(str); err != nil {
			gui.GUI.OutputWrite("Invalid Address")
			continue
		}
		break
	}

	if amount, err = builder.readAmount(assetId, text+" Amount"); err != nil {
		return
	}

	addressEncoded = address.EncodeAddr()
	return
}

func (builder *TxsBuilder) readZetherRingConfiguration() *ZetherRingConfiguration {

	configuration := &ZetherRingConfiguration{
		-1, &ZetherSenderRingType{}, &ZetherRecipientRingType{},
	}
	configuration.RingSize = gui.GUI.OutputReadInt("Ring Size (2,4,8,16,32,64,128,256). Leave empty for random", true, -1, func(value int) bool {
		switch value {
		case 2, 4, 8, 16, 32, 64, 128, 256:
			return true
		default:
			return false
		}
	})

	configuration.RecipientRingType.NewAccounts = gui.GUI.OutputReadInt("Ring New Accounts (0...n-2). Use empty for random", true, -1, func(value int) bool {
		return value >= 0
	})

	return configuration
}

func (builder *TxsBuilder) readFee(assetId []byte) (fee *wizard.WizardTransactionFee) {

	var err error
	fee = &wizard.WizardTransactionFee{}

	fee.PerByteAuto = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte? y/n. Leave empty for yes", true, true)
	if !fee.PerByteAuto {

		if fee.PerByte, err = builder.readAmount(assetId, "Fee per byte"); err != nil {
			panic(err)
		}

		if fee.PerByte == 0 {
			if fee.Fixed, err = builder.readAmount(assetId, "Fee per byte"); err != nil {
				panic(err)
			}
		}
	}

	return
}

func (builder *TxsBuilder) readZetherFee(assetId []byte) (fee *wizard.WizardZetherTransactionFee) {

	fee = &wizard.WizardZetherTransactionFee{}
	fee.WizardTransactionFee = builder.readFee(assetId)

	if !bytes.Equal(assetId, config_coins.NATIVE_ASSET_FULL) {
		fee.Auto = gui.GUI.OutputReadBool("Compute autoamtically Fee Rate Max for Asset. y/n. Leave empty for yes", true, true)
		if !fee.Auto {
			fee.Rate = gui.GUI.OutputReadUint64("Fee Rate for Asset", true, 0, nil)
			fee.LeadingZeros = byte(gui.GUI.OutputReadUint64("Fee Leading Zeros for Asset", true, 0, func(value uint64) bool {
				return value <= uint64(config_assets.ASSETS_DECIMAL_SEPARATOR_MAX)
			}))
		}
	}

	return
}

func (builder *TxsBuilder) readAsset(text string, allowEmptyAsset bool) []byte {
	assetId := gui.GUI.OutputReadBytes(text, func(input []byte) bool {
		return (allowEmptyAsset && len(input) == 0) || len(input) == config_coins.ASSET_LENGTH
	})
	if len(assetId) == 0 {
		assetId = config_coins.NATIVE_ASSET_FULL
	}
	return assetId
}

func (builder *TxsBuilder) initCLI() {

	cliPrivateTransfer := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		txData := &TxBuilderCreateZetherTxData{
			Payloads: []*TxBuilderCreateZetherTxPayload{{}},
		}

		if _, txData.Payloads[0].Sender, _, err = builder.wallet.CliSelectAddress("Select Address to Transfer", ctx); err != nil {
			return
		}

		txData.Payloads[0].Asset = builder.readAsset("Asset. Leave empty for Native Asset", true)

		if _, txData.Payloads[0].Recipient, txData.Payloads[0].Amount, err = builder.readAddressOptional("Recipient Address", txData.Payloads[0].Asset, false); err != nil {
			return
		}

		txData.Payloads[0].RingConfiguration = builder.readZetherRingConfiguration()
		txData.Payloads[0].Data = builder.readData()
		txData.Payloads[0].Fee = builder.readZetherFee(txData.Payloads[0].Asset)
		propagate := gui.GUI.OutputReadBool("Propagate? y/n. Leave empty for yes", true, true)

		tx, err := builder.CreateZetherTx(txData, nil, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Tx created: %s %s", base64.StdEncoding.EncodeToString(tx.Bloom.Hash), cmd))
		return
	}

	cliPrivateAssetCreate := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		extra := &wizard.WizardZetherPayloadExtraAssetCreate{}
		txData := &TxBuilderCreateZetherTxData{
			Payloads: []*TxBuilderCreateZetherTxPayload{{
				Extra: extra,
				Asset: config_coins.NATIVE_ASSET_FULL,
			}},
		}

		if _, txData.Payloads[0].Sender, _, err = builder.wallet.CliSelectAddress("Select Address which will create the asset", ctx); err != nil {
			return
		}

		extra.Asset = asset.NewAsset(nil, 0)
		str := gui.GUI.OutputReadString("Asset as JSON")
		if err = json.Unmarshal([]byte(str), extra.Asset); err != nil {
			return
		}
		extra.Asset.PublicKeyHash = helpers.RandomBytes(cryptography.PublicKeyHashSize)
		extra.Asset.Identification = extra.Asset.Ticker + "-" + hex.EncodeToString(extra.Asset.PublicKeyHash[:3])

		if err = extra.Asset.Validate(); err != nil {
			return
		}

		var updatePrivKey, supplyPrivKey *addresses.PrivateKey
		if len(extra.Asset.UpdatePublicKey) == 0 {
			updatePrivKey = addresses.GenerateNewPrivateKey()
			extra.Asset.UpdatePublicKey = updatePrivKey.GeneratePublicKey()
		}
		if len(extra.Asset.SupplyPublicKey) == 0 {
			supplyPrivKey = addresses.GenerateNewPrivateKey()
			extra.Asset.SupplyPublicKey = supplyPrivKey.GeneratePublicKey()
		}

		if _, txData.Payloads[0].Recipient, txData.Payloads[0].Amount, err = builder.readAddressOptional("Transfer Address", config_coins.NATIVE_ASSET_FULL, true); err != nil {
			return
		}

		txData.Payloads[0].RingConfiguration = builder.readZetherRingConfiguration()
		txData.Payloads[0].Data = builder.readData()
		txData.Payloads[0].Fee = builder.readZetherFee(config_coins.NATIVE_ASSET_FULL)

		propagate := gui.GUI.OutputReadBool("Propagate? y/n. Leave empty for yes", true, true)

		tx, err := builder.CreateZetherTx(txData, nil, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Tx created: %s %s", base64.StdEncoding.EncodeToString(tx.Bloom.Hash), cmd))

		assetId := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).Payloads[0].Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate).GetAssetId(tx.Bloom.Hash, 0)
		gui.GUI.OutputWrite(fmt.Sprintf("Asset Id: %s", base64.StdEncoding.EncodeToString(assetId), cmd))

		if updatePrivKey != nil || supplyPrivKey != nil {

			if filename := gui.GUI.OutputReadFilename("Path to export Asset Private Keys", "keys", true); len(filename) > 0 {
				if err = files.WriteFile(filename,
					fmt.Sprintf("Asset ID: %s", base64.StdEncoding.EncodeToString(assetId)),
					fmt.Sprintf("Asset name: %s", extra.Asset.Name, extra.Asset.Ticker),
					fmt.Sprintf("Supply Private Key: %s", base64.StdEncoding.EncodeToString(supplyPrivKey.Key)),
					fmt.Sprintf("Update Private Key: %s", base64.StdEncoding.EncodeToString(updatePrivKey.Key)),
				); err != nil {
					return
				}
				gui.GUI.OutputWrite("Asset Keys Exported successfully to: ", filename)
			}

		}

		return
	}

	cliPrivateAssetSupplyIncrease := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		extra := &wizard.WizardZetherPayloadExtraAssetSupplyIncrease{}
		txData := &TxBuilderCreateZetherTxData{
			Payloads: []*TxBuilderCreateZetherTxPayload{{
				Extra: extra,
				Asset: config_coins.NATIVE_ASSET_FULL,
			}},
		}

		if _, txData.Payloads[0].Sender, _, err = builder.wallet.CliSelectAddress("Select Address which will increase the supply of asset", ctx); err != nil {
			return
		}

		extra.AssetId = builder.readAsset("Asset", false)

		extra.AssetSupplyPrivateKey = gui.GUI.OutputReadBytes("Asset Supply Update Private Key", func(value []byte) bool {
			return len(value) == cryptography.PrivateKeySize
		})

		var receiverAddress *addresses.Address
		if receiverAddress, _, extra.Value, err = builder.readAddressOptional("Receiver Address", extra.AssetId, false); err != nil {
			return
		}
		extra.ReceiverPublicKey = receiverAddress.PublicKey

		if _, txData.Payloads[0].Recipient, txData.Payloads[0].Amount, err = builder.readAddressOptional("Transfer Address", config_coins.NATIVE_ASSET_FULL, true); err != nil {
			return
		}

		txData.Payloads[0].RingConfiguration = builder.readZetherRingConfiguration()
		txData.Payloads[0].Data = builder.readData()
		txData.Payloads[0].Fee = builder.readZetherFee(config_coins.NATIVE_ASSET_FULL)
		propagate := gui.GUI.OutputReadBool("Propagate? y/n. Leave empty for yes", true, true)

		tx, err := builder.CreateZetherTx(txData, nil, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Tx created: %s %s", base64.StdEncoding.EncodeToString(tx.Bloom.Hash), cmd))

		return
	}

	cliPrivatePlainAccountFund := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		extra := &wizard.WizardZetherPayloadExtraPlainAccountFund{}
		txData := &TxBuilderCreateZetherTxData{
			Payloads: []*TxBuilderCreateZetherTxPayload{{
				Extra: extra,
				Asset: config_coins.NATIVE_ASSET_FULL,
			}},
		}

		if _, txData.Payloads[0].Sender, _, err = builder.wallet.CliSelectAddress("Select Address which will fund a plain account", ctx); err != nil {
			return
		}

		var plainAccountAddress *addresses.Address
		if plainAccountAddress, _, txData.Payloads[0].Burn, err = builder.readAddressOptional("Plain Account", config_coins.NATIVE_ASSET_FULL, false); err != nil {
			return
		}

		extra.PlainAccountPublicKey = plainAccountAddress.PublicKey

		if _, txData.Payloads[0].Recipient, txData.Payloads[0].Amount, err = builder.readAddressOptional("Transfer Address", config_coins.NATIVE_ASSET_FULL, true); err != nil {
			return
		}

		txData.Payloads[0].RingConfiguration = builder.readZetherRingConfiguration()
		txData.Payloads[0].Data = builder.readData()
		txData.Payloads[0].Fee = builder.readZetherFee(config_coins.NATIVE_ASSET_FULL)

		propagate := gui.GUI.OutputReadBool("Propagate? y/n. Leave empty for yes", true, true)

		tx, err := builder.CreateZetherTx(txData, nil, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Tx created: %s %s", base64.StdEncoding.EncodeToString(tx.Bloom.Hash), cmd))

		return
	}

	cliPrivatePayInFuture := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		extra := &wizard.WizardZetherPayloadExtraPayInFuture{}
		txData := &TxBuilderCreateZetherTxData{
			Payloads: []*TxBuilderCreateZetherTxPayload{{
				Extra: extra,
			}, {}},
		}

		if _, txData.Payloads[0].Sender, _, err = builder.wallet.CliSelectAddress("Select Address to Transfer", ctx); err != nil {
			return
		}
		txData.Payloads[1].Sender = txData.Payloads[0].Sender

		txData.Payloads[0].Asset = builder.readAsset("Asset. Leave empty for Native Asset", true)
		txData.Payloads[1].Asset = txData.Payloads[0].Asset

		if _, txData.Payloads[0].Recipient, txData.Payloads[0].Amount, err = builder.readAddressOptional("Recipient Address", txData.Payloads[0].Asset, false); err != nil {
			return
		}

		extra.Deadline = gui.GUI.OutputReadUint64("Deadline", true, 10, func(val uint64) bool {
			return val >= 10 && val <= 100000
		})

		extra.DefaultResolution = gui.GUI.OutputReadBool("Default Resolution: y - reciever, n - sender", false, false)

		extra.Threshold = byte(gui.GUI.OutputReadUint64("Threshold", true, 1, func(val uint64) bool {
			return val >= 1 && val <= 5
		}))

		extra.MultisigPublicKeys = [][]byte{}
		unique := make(map[string]bool)
		for {
			pubKey := gui.GUI.OutputReadBytes(fmt.Sprintf("PublicKey %d used in multisig payment", len(extra.MultisigPublicKeys)), func(val []byte) bool {
				return len(val) == 0 || len(val) == cryptography.PublicKeySize
			})
			if len(pubKey) == 0 {
				break
			}
			if unique[string(pubKey)] {
				gui.GUI.OutputWrite("PublicKey already include")
				continue
			}
			extra.MultisigPublicKeys = append(extra.MultisigPublicKeys, pubKey)
		}

		if _, txData.Payloads[1].Recipient, txData.Payloads[1].Amount, err = builder.readAddressOptional("Transfer Address (optional)", config_coins.NATIVE_ASSET_FULL, true); err != nil {
			return
		}

		txData.Payloads[0].RingConfiguration = builder.readZetherRingConfiguration()
		if err = builder.presetZetherRing(txData.Payloads[0].RingConfiguration); err != nil {
			return err
		}

		txData.Payloads[0].RingConfiguration.SenderRingType.AvoidStakedAccounts = true
		txData.Payloads[0].RingConfiguration.RecipientRingType.AvoidStakedAccounts = true

		txData.Payloads[1].RingConfiguration = &ZetherRingConfiguration{
			txData.Payloads[0].RingConfiguration.RingSize,
			&ZetherSenderRingType{false, true, []string{}, 0},
			&ZetherRecipientRingType{false, true, []string{}, txData.Payloads[0].RingConfiguration.RecipientRingType.NewAccounts},
		}

		txData.Payloads[0].Data = builder.readData()

		txData.Payloads[0].Fee = builder.readZetherFee(txData.Payloads[0].Asset)
		txData.Payloads[1].Fee = txData.Payloads[0].Fee
		propagate := gui.GUI.OutputReadBool("Propagate? y/n. Leave empty for yes", true, true)

		tx, err := builder.CreateZetherTx(txData, nil, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Tx created: %s %s", base64.StdEncoding.EncodeToString(tx.Bloom.Hash), cmd))
		return
	}

	cliUpdateAssetFeeLiquidity := func(cmd string, ctx context.Context) (err error) {

		builder.showWarningIfNotSyncCLI()

		txExtra := &wizard.WizardTxSimpleExtraUpdateAssetFeeLiquidity{}
		txData := &TxBuilderCreateSimpleTx{
			Extra:      txExtra,
			FeeVersion: true,
		}

		if _, txData.Sender, _, err = builder.wallet.CliSelectAddress("Select Address to Publicly Update Asset Fee Liquidity", ctx); err != nil {
			return
		}

		var addr *addresses.Address
		if addr, err = builder.readAddress("Collector address. Leave empty for no new address", true); err != nil {
			return
		}
		if addr != nil {
			txExtra.NewCollector = true
			txExtra.Collector = addr.PublicKey
		}

		for {
			newLiquidity := gui.GUI.OutputReadBool("New Liquidity? y/n", false, false)
			if !newLiquidity {
				break
			}
			liquidity := &asset_fee_liquidity.AssetFeeLiquidity{}
			liquidity.Asset = builder.readAsset("Asset", false)
			liquidity.Rate = gui.GUI.OutputReadUint64("Conversion Rate", false, 0, nil)
			liquidity.LeadingZeros = byte(gui.GUI.OutputReadUint64("Leading Zeros", true, 0, func(value uint64) bool {
				return value <= uint64(config_assets.ASSETS_DECIMAL_SEPARATOR_MAX_BYTE)
			}))
			txExtra.Liquidities = append(txExtra.Liquidities, liquidity)
		}

		txData.Nonce = gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", true, 0, nil)
		txData.Data = builder.readData()
		txData.Fee = builder.readFee(config_coins.NATIVE_ASSET_FULL)

		propagate := gui.GUI.OutputReadBool("Propagate? y/n. Leave empty for yes", true, true)

		tx, err := builder.CreateSimpleTx(txData, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Tx created: %s %s", base64.StdEncoding.EncodeToString(tx.Bloom.Hash), cmd))
		return
	}

	cliResolutionPayInFuture := func(cmd string, ctx context.Context) (err error) {

		builder.showWarningIfNotSyncCLI()

		txExtra := &wizard.WizardTxSimpleExtraResolutionPayInFuture{
			MultisigPublicKeys: make([][]byte, 0),
			Signatures:         make([][]byte, 0),
		}
		txData := &TxBuilderCreateSimpleTx{
			Extra:      txExtra,
			Fee:        &wizard.WizardTransactionFee{0, 0, 0, false},
			FeeVersion: true,
		}

		txExtra.TxId = gui.GUI.OutputReadBytes("Provide TxId", func(val []byte) bool {
			return len(val) == cryptography.HashSize
		})

		txExtra.PayloadIndex = byte(gui.GUI.OutputReadInt("Payload index", false, 0, func(val int) bool {
			return val >= 0 && val < 255
		}))

		txExtra.Resolution = gui.GUI.OutputReadBool("Resolution.  Use y/n for voting", false, false)

		i := 0
		for {
			key := gui.GUI.OutputReadBytes(fmt.Sprintf("Public Key %d. Use enter to continue", i), func(key []byte) bool {
				return len(key) == cryptography.PublicKeySize || len(key) == 0
			})
			if len(key) == 0 {
				break
			}

			signature := gui.GUI.OutputReadBytes("Signature", func(sign []byte) bool {
				return len(sign) == cryptography.SignatureSize
			})

			extra := &transaction_simple_extra.TransactionSimpleExtraResolutionPayInFuture{nil,
				txExtra.TxId,
				txExtra.PayloadIndex,
				txExtra.Resolution,
				[][]byte{key},
				[][]byte{signature},
			}

			if !extra.VerifySignature() {
				gui.GUI.Error("provided resolution signature is not valid")
				break
			}

			txExtra.MultisigPublicKeys = append(txExtra.MultisigPublicKeys, key)
			txExtra.Signatures = append(txExtra.Signatures, signature)

			i++
		}

		txData.Nonce = 0
		txData.Data = builder.readData()

		propagate := gui.GUI.OutputReadBool("Propagate? y/n. Leave empty for yes", true, true)

		tx, err := builder.CreateSimpleTx(txData, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite(fmt.Sprintf("Tx created: %s %s", base64.StdEncoding.EncodeToString(tx.Bloom.Hash), cmd))
		return
	}

	gui.GUI.CommandDefineCallback("Private Transfer", cliPrivateTransfer, true)
	gui.GUI.CommandDefineCallback("Private Asset Create", cliPrivateAssetCreate, true)
	gui.GUI.CommandDefineCallback("Private Asset Supply Increase", cliPrivateAssetSupplyIncrease, true)
	gui.GUI.CommandDefineCallback("Private Plain Account Fund", cliPrivatePlainAccountFund, true)
	gui.GUI.CommandDefineCallback("Private Pay In Future", cliPrivatePayInFuture, true)
	gui.GUI.CommandDefineCallback("Public Update Asset Fee Liquidity", cliUpdateAssetFeeLiquidity, true)
	gui.GUI.CommandDefineCallback("Public Resolution Pay in Future", cliResolutionPayInFuture, true)

}
