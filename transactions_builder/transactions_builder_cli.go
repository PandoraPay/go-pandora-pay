package transactions_builder

import (
	"context"
	"encoding/hex"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/transactions_builder/wizard"
)

func (builder *TransactionsBuilder) showWarningIfNotSyncCLI() {
	if builder.chain.Sync.GetSyncTime() == 0 {
		gui.GUI.OutputWrite("Your node is not Sync yet. Wait for it to get sync.")
	}
}

func (builder *TransactionsBuilder) readData() (out *wizard.TransactionsWizardData) {

	data := &wizard.TransactionsWizardData{}
	str := gui.GUI.OutputReadString("Message (data). Leave empty for none")

	if len(str) > 0 {
		data.Data = []byte(str)
		data.Encrypt = gui.GUI.OutputReadBool("Encrypt message (data)? y/n")
	}

	return data
}

func (builder *TransactionsBuilder) readAmount(assetId []byte, text string) (amount uint64, err error) {

	amountFloat := gui.GUI.OutputReadFloat64(text, false, nil)

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

func (builder *TransactionsBuilder) readAddress(text string, assetId []byte, allowRandomAddress bool) (address *addresses.Address, amount uint64, err error) {

	text2 := text
	if allowRandomAddress {
		text2 = text + ". Leave empty for none"
	}

	address = gui.GUI.OutputReadAddress(text2, allowRandomAddress)
	if address == nil {
		return
	} else {
		if amount, err = builder.readAmount(assetId, text+" Amount"); err != nil {
			return
		}
	}

	return
}

func (builder *TransactionsBuilder) readZetherRingConfiguration() *ZetherRingConfiguration {

	configuration := &ZetherRingConfiguration{}
	configuration.RingSize = gui.GUI.OutputReadInt("Ring Size (-1,2,4,8,16,32,64,128,256):", false, func(value int) bool {
		switch value {
		case -1, 2, 4, 8, 16, 32, 64, 128, 256:
			return true
		default:
			return false
		}
	})

	configuration.NewAccounts = gui.GUI.OutputReadInt("Ring New Accounts (-1,0...n-2)", false, func(value int) bool {
		return value >= -1 && value <= configuration.RingSize-2
	})

	return configuration
}

func (builder *TransactionsBuilder) readFees(assetId []byte) (fee *wizard.TransactionsWizardFee, err error) {

	fee = &wizard.TransactionsWizardFee{}

	fee.PerByteAuto = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte? y/n")
	if !fee.PerByteAuto {

		if fee.PerByte, err = builder.readAmount(assetId, "Fee per byte"); err != nil {
			return
		}

		if fee.PerByte == 0 {
			if fee.Fixed, err = builder.readAmount(assetId, "Fee per byte"); err != nil {
				return
			}
		}
	}

	return
}

func (builder *TransactionsBuilder) readDelegatedStakingUpdate(delegatedStakingUpdate *transaction_data.TransactionDataDelegatedStakingUpdate, delegateWalletPublicKey []byte) (err error) {
	delegatedStakingUpdate.DelegatedStakingHasNewInfo = gui.GUI.OutputReadBool("New Delegate Info? y/n")

	if delegatedStakingUpdate.DelegatedStakingHasNewInfo {

		delegatedStakingUpdate.DelegatedStakingNewPublicKey = gui.GUI.OutputReadBytes("Delegated Staking New PublicKey. Leave Empty to automatically derive pubKey", func(input []byte) bool {
			return len(input) == 0 || len(input) == cryptography.PublicKeySize
		})

		if len(delegatedStakingUpdate.DelegatedStakingNewPublicKey) == 0 {
			if delegatedStakingUpdate.DelegatedStakingNewPublicKey, _, err = builder.DeriveDelegatedStake(0, delegateWalletPublicKey); err != nil {
				return
			}
		}

		delegatedStakingUpdate.DelegatedStakingNewFee = gui.GUI.OutputReadUint64("Delegated Staking New Fee. Leave empty for nothing", true, func(value uint64) bool {
			return value <= config_stake.DELEGATING_STAKING_FEES_MAX_VALUE
		})
	}

	return
}

func (builder *TransactionsBuilder) initCLI() {

	cliPrivateTransfer := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Transfer", ctx)
		if err != nil {
			return
		}

		assetId := gui.GUI.OutputReadBytes("Asset. Leave empty for Native Asset", func(input []byte) bool {
			return len(input) == 0 || len(input) == config_coins.ASSET_LENGTH
		})
		if len(assetId) == 0 {
			assetId = config_coins.NATIVE_ASSET_FULL
		}

		destinationAddress, amount, err := builder.readAddress("Destination Address", assetId, false)
		if err != nil {
			return
		}

		ringConfiguration := builder.readZetherRingConfiguration()

		data := builder.readData()

		fee, err := builder.readFees(assetId)
		if err != nil {
			return
		}

		propagate := gui.GUI.OutputReadBool("Propagate? y/n")

		tx, err := builder.CreateZetherTx([]wizard.ZetherTransferPayloadExtra{nil}, []string{walletAddress.AddressEncoded}, [][]byte{assetId}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{0}, []*ZetherRingConfiguration{ringConfiguration}, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliPrivateDelegateStake := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address from which Delegate", ctx)
		if err != nil {
			return
		}

		delegateAddress, delegateAmount, err := builder.readAddress("Delegate Address", config_coins.NATIVE_ASSET_FULL, false)
		if err != nil {
			return
		}

		var delegatePrivateKey []byte

		delegateWalletAddress := builder.wallet.GetWalletAddressByPublicKey(delegateAddress.PublicKey)
		delegatedStakingUpdate := &transaction_data.TransactionDataDelegatedStakingUpdate{}
		if delegateWalletAddress != nil {
			if err = builder.readDelegatedStakingUpdate(delegatedStakingUpdate, delegateWalletAddress.PublicKey); err != nil {
				return
			}
		}
		if delegatedStakingUpdate.DelegatedStakingHasNewInfo {
			delegatePrivateKey = delegateWalletAddress.PrivateKey.Key
		}

		destinationAddress, destinationAmount, err := builder.readAddress("Transfer Destination Address", config_coins.NATIVE_ASSET_FULL, true)
		if err != nil {
			return
		}

		ringConfiguration := builder.readZetherRingConfiguration()

		data := builder.readData()

		fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}

		propagate := gui.GUI.OutputReadBool("Propagate? y/n")

		tx, err := builder.CreateZetherTx([]wizard.ZetherTransferPayloadExtra{&wizard.ZetherTransferPayloadExtraDelegateStake{DelegatePublicKey: delegateAddress.PublicKey, DelegatedStakingUpdate: delegatedStakingUpdate, DelegatePrivateKey: delegatePrivateKey}}, []string{walletAddress.AddressEncoded}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{destinationAmount}, []string{destinationAddress.EncodeAddr()}, []uint64{delegateAmount}, []*ZetherRingConfiguration{ringConfiguration}, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliPrivateClaimStake := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		delegateWalletAddress, _, err := builder.wallet.CliSelectAddress("Select Address from which Claim", ctx)
		if err != nil {
			return
		}

		destinationAddress, amount, err := builder.readAddress("Claim Address", config_coins.NATIVE_ASSET_FULL, false)
		if err != nil {
			return
		}

		ringConfiguration := builder.readZetherRingConfiguration()

		data := builder.readData()

		fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}

		propagate := gui.GUI.OutputReadBool("Propagate? y/n")

		tx, err := builder.CreateZetherTx([]wizard.ZetherTransferPayloadExtra{&wizard.ZetherTransferPayloadExtraClaimStake{DelegatePrivateKey: delegateWalletAddress.PrivateKey.Key}}, []string{""}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{0}, []*ZetherRingConfiguration{ringConfiguration}, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliPrivateAssetCreate := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()
		//
		//walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address which will create the asset", ctx)
		//if err != nil {
		//	return
		//}
		//
		//destinationAddress, amount, err := builder.readAddress("Destination Address", config_coins.NATIVE_ASSET_FULL, true )
		//
		//ringConfiguration := builder.readZetherRingConfiguration()
		//
		//data := builder.readData()
		//
		//fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		//if err != nil {
		//	return
		//}
		//
		//propagate := gui.GUI.OutputReadBool("Propagate? y/n")
		//
		//tx, err := builder.CreateZetherTx([]wizard.ZetherTransferPayloadExtra{&wizard.ZetherTransferPayloadExtraAssetCreate{ Asset: ast }}, []string{""}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{0}, []*ZetherRingConfiguration{ringConfiguration}, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
		//	gui.GUI.OutputWrite(status)
		//})
		//if err != nil {
		//	return
		//}
		//
		//gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		return
	}

	cliPrivateAssetSupplyIncrease := func(cmd string, ctx context.Context) (err error) {
		builder.showWarningIfNotSyncCLI()

		return
	}

	cliUpdateDelegate := func(cmd string, ctx context.Context) (err error) {

		builder.showWarningIfNotSyncCLI()

		delegateWalletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Update Delegate", ctx)
		if err != nil {
			return
		}

		nonce := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", true, nil)

		delegatedStakingUpdate := &transaction_data.TransactionDataDelegatedStakingUpdate{}
		if err = builder.readDelegatedStakingUpdate(delegatedStakingUpdate, delegateWalletAddress.PublicKey); err != nil {
			return
		}

		delegatedStakingClaimAmount, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Update Delegated Staking Amount")
		if err != nil {
			return
		}

		data := builder.readData()

		fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}

		propagate := gui.GUI.OutputReadBool("Propagate? y/n")

		tx, err := builder.CreateUpdateDelegateTx(delegateWalletAddress.AddressEncoded, nonce, delegatedStakingClaimAmount, delegatedStakingUpdate, data, fee, propagate, true, true, false, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliUnstake := func(cmd string, ctx context.Context) (err error) {

		builder.showWarningIfNotSyncCLI()

		delegateWalletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Unstake", ctx)
		if err != nil {
			return
		}

		amount, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Amount")
		if err != nil {
			return
		}

		nonce := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", true, nil)

		data := builder.readData()

		fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}

		propagate := gui.GUI.OutputReadBool("Propagate? y/n")

		tx, err := builder.CreateUnstakeTx(delegateWalletAddress.AddressEncoded, nonce, amount, data, fee, propagate, true, true, false, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	gui.GUI.CommandDefineCallback("Private Transfer", cliPrivateTransfer, true)
	gui.GUI.CommandDefineCallback("Private Delegate Stake", cliPrivateDelegateStake, true)
	gui.GUI.CommandDefineCallback("Private Claim Stake", cliPrivateClaimStake, true)
	gui.GUI.CommandDefineCallback("Private Asset Create", cliPrivateAssetCreate, true)
	gui.GUI.CommandDefineCallback("Private Asset Supply Increase", cliPrivateAssetSupplyIncrease, true)
	gui.GUI.CommandDefineCallback("Update Delegate", cliUpdateDelegate, true)
	gui.GUI.CommandDefineCallback("Unstake", cliUnstake, true)

}
