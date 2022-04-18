package txs_builder

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/gui"
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
		if ast, err = asts.GetAsset(assetId); err != nil {
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

		address, err = addresses.DecodeAddr(str)
		if err != nil {
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

}
