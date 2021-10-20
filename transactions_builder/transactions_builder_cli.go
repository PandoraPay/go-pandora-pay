package transactions_builder

import (
	"context"
	"encoding/hex"
	"errors"
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
		data.Encrypt = gui.GUI.OutputReadBool("Encrypt message (data)? [y/n]")
	}

	return data
}

func (builder *TransactionsBuilder) readAmount(assetId []byte, text string) (amount uint64, err error) {

	amountFloat := gui.GUI.OutputReadFloat64(text, nil)

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

func (builder *TransactionsBuilder) readFees(assetId []byte) (fee *wizard.TransactionsWizardFee, err error) {

	fee = &wizard.TransactionsWizardFee{}

	fee.PerByteAuto = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte? [y/n]")
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
	delegatedStakingUpdate.DelegatedStakingHasNewInfo = gui.GUI.OutputReadBool("New Delegate Info? [y/n]")

	if delegatedStakingUpdate.DelegatedStakingHasNewInfo {

		delegatedStakingUpdate.DelegatedStakingNewPublicKey = gui.GUI.OutputReadBytes("Delegated Staking New PublicKey. Leave Empty to automatically derive pubKey", func(input []byte) bool {
			return len(input) == 0 || len(input) == cryptography.PublicKeySize
		})

		if len(delegatedStakingUpdate.DelegatedStakingNewPublicKey) == 0 {
			if delegatedStakingUpdate.DelegatedStakingNewPublicKey, _, err = builder.DeriveDelegatedStake(0, delegateWalletPublicKey); err != nil {
				return
			}
		}

		delegatedStakingUpdate.DelegatedStakingNewFee = gui.GUI.OutputReadUint64("Delegated Staking New Fee. Leave empty for nothing", func(value uint64) bool {
			return value <= config_stake.DELEGATING_STAKING_FEES_MAX_VALUE
		})
	}

	return
}

func (builder *TransactionsBuilder) initCLI() {

	cliPrivateTransfer := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Transfer")
		if err != nil {
			return
		}

		assetId := gui.GUI.OutputReadBytes("Asset. Leave empty for Native Asset", func(input []byte) bool {
			return len(input) == 0 || len(input) == config_coins.ASSET_LENGTH
		})
		if len(assetId) == 0 {
			assetId = config_coins.NATIVE_ASSET_FULL
		}

		amount, err := builder.readAmount(assetId, "Amount")
		if err != nil {
			return
		}

		destinationAddress := gui.GUI.OutputReadAddress("Destination Address")
		data := builder.readData()
		propagate := gui.GUI.OutputReadBool("Propagate? [y/n]")

		fee, err := builder.readFees(assetId)
		if err != nil {
			return
		}

		ringMembers := make([][]string, 1)
		if ringMembers[0], err = builder.CreateZetherRing(walletAddress.AddressEncoded, destinationAddress.EncodeAddr(), assetId, -1, -1); err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tx, err := builder.CreateZetherTx([]string{walletAddress.AddressEncoded}, [][]byte{assetId}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{0}, ringMembers, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliPrivateDelegateStake := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address from which to Delegate")
		if err != nil {
			return
		}

		delegateAmount, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Delegate Amount")
		if err != nil {
			return
		}

		delegateAddress := gui.GUI.OutputReadAddress("Delegate Address")

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

		destinationAddress := gui.GUI.OutputReadAddress("Destination Address")

		amount, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Amount")
		if err != nil {
			return
		}

		data := builder.readData()

		propagate := gui.GUI.OutputReadBool("Propagate? [y/n]")

		fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}

		ringMembers := make([][]string, 1)
		if ringMembers[0], err = builder.CreateZetherRing(walletAddress.AddressEncoded, destinationAddress.EncodeAddr(), config_coins.NATIVE_ASSET_FULL, -1, -1); err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tx, err := builder.CreateZetherDelegateStakeTx(delegateAddress.PublicKey, delegatedStakingUpdate, delegatePrivateKey, []string{walletAddress.AddressEncoded}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{delegateAmount}, ringMembers, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliPrivateClaimStake := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		delegateWalletAddress, _, err := builder.wallet.CliSelectAddress("Select Address from which should Claim")
		if err != nil {
			return
		}

		amount, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Amount to Claim")
		if err != nil {
			return
		}

		destinationAddress := gui.GUI.OutputReadAddress("Destination Address")

		data := builder.readData()

		propagate := gui.GUI.OutputReadBool("Propagate? [y/n]")

		fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}

		ringMembers := make([][]string, 1)
		if ringMembers[0], err = builder.CreateZetherRing("", destinationAddress.EncodeAddr(), config_coins.NATIVE_ASSET_FULL, -1, -1); err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tx, err := builder.CreateZetherClaimStakeTx(delegateWalletAddress.PrivateKey.Key, []string{""}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{0}, ringMembers, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliUpdateDelegate := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		delegateWalletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Update Delegate")
		if err != nil {
			return
		}

		nonce := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", nil)

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

		propagate := gui.GUI.OutputReadBool("Propagate? [y/n]")

		tx, err := builder.CreateUpdateDelegateTx(delegateWalletAddress.AddressEncoded, nonce, delegatedStakingClaimAmount, delegatedStakingUpdate, data, fee, propagate, true, true, false, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	cliUnstake := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		delegateWalletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Amount")
		if err != nil {
			return
		}

		nonce := gui.GUI.OutputReadUint64("Nonce. Leave for automatically detection", nil)

		data := builder.readData()

		fee, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}

		propagate := gui.GUI.OutputReadBool("Propagate? [y/n]")

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
	gui.GUI.CommandDefineCallback("Update Delegate", cliUpdateDelegate, true)
	gui.GUI.CommandDefineCallback("Unstake", cliUnstake, true)

}
