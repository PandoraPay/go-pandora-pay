package transactions_builder

import (
	"context"
	"encoding/hex"
	"errors"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/gui"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/wallet/wallet_address"
)

func (builder *TransactionsBuilder) showWarningIfNotSyncCLI() {
	if builder.chain.Sync.GetSyncTime() == 0 {
		gui.GUI.OutputWrite("Your node is not Sync yet. Wait for it to get sync.")
	}
}

func (builder *TransactionsBuilder) readData() (out *wizard.TransactionsWizardData, ok bool) {

	data := &wizard.TransactionsWizardData{}
	str, ok := gui.GUI.OutputReadString("Message (data). Leave empty for none")
	if !ok {
		return
	}

	if len(str) > 0 {
		data.Data = []byte(str)

		data.Encrypt, ok = gui.GUI.OutputReadBool("Encrypt message (data). Type y/n")

	}

	return data, ok
}

func (builder *TransactionsBuilder) readAmount(assetId []byte, text string) (amount uint64, ok bool, err error) {

	amountFloat, ok := gui.GUI.OutputReadFloat64(text, nil)
	if !ok {
		return
	}

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

func (builder *TransactionsBuilder) readFees(assetId []byte) (fee *wizard.TransactionsWizardFee, ok bool, err error) {

	fee = &wizard.TransactionsWizardFee{}

	if fee.PerByteAuto, ok = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte. Type y\n"); !ok {
		return
	}
	if !fee.PerByteAuto {

		if fee.PerByte, ok, err = builder.readAmount(assetId, "Fee per byte"); !ok || err != nil {
			return
		}

		if fee.PerByte == 0 {
			if fee.Fixed, ok, err = builder.readAmount(assetId, "Fee per byte"); !ok || err != nil {
				return
			}
		}
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

		assetId, ok := gui.GUI.OutputReadBytes("Asset. Leave empty for Native Asset", []int{config_coins.ASSET_LENGTH})
		if !ok {
			return
		}
		if len(assetId) != 40 {
			return errors.New("Invalid AssetId")
		}

		amount, ok, err := builder.readAmount(assetId, "Amount")
		if !ok || err != nil {
			return
		}

		destinationAddress, ok := gui.GUI.OutputReadAddress("Destination Address")
		if !ok {
			return
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		fee, ok, err := builder.readFees(assetId)
		if !ok || err != nil {
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

		delegateAmount, ok, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Delegate Amount")
		if !ok || err != nil {
			return
		}

		delegateAddress, ok := gui.GUI.OutputReadAddress("Delegate Address")
		if !ok {
			return
		}

		delegatedStakingHasNewInfo := false
		var delegatePrivateKey, delegatedStakingNewPublicKey []byte
		var delegatedStakingNewFee uint64

		delegateWalletAddress := builder.wallet.GetWalletAddressByPublicKey(delegateAddress.PublicKey)
		if delegateWalletAddress != nil {

			if delegatedStakingHasNewInfo, ok = gui.GUI.OutputReadBool("New Delegate Info ? Type y/n"); !ok {
				return
			}

			if delegatedStakingHasNewInfo {

				delegatePrivateKey = delegateWalletAddress.PrivateKey.Key

				if delegatedStakingNewPublicKey, ok = gui.GUI.OutputReadBytes("Delegated Staking New PublicKey. Leave Empty to automatically derive pubKey", []int{20, 0}); !ok {
					return
				}

				if len(delegatedStakingNewPublicKey) == 0 {
					var derivedDelegatedStake *wallet_address.WalletAddressDelegatedStake
					if derivedDelegatedStake, err = builder.DeriveDelegatedStake(0, walletAddress.PublicKey); err != nil {
						return
					}
					delegatedStakingNewPublicKey = derivedDelegatedStake.PublicKey
				}

				delegatedStakingNewFee, ok = gui.GUI.OutputReadUint64("Delegated Staking New Fee. Leave empty for nothing", nil, true)
				if delegatedStakingNewFee > config_stake.DELEGATING_STAKING_FEES_MAX_VALUE {
					return errors.New("Invalid NewFee")
				}

			}

		}

		destinationAddress, ok := gui.GUI.OutputReadAddress("Destination Address")
		if !ok {
			return
		}

		amount, ok, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Amount")
		if !ok || err != nil {
			return
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		fee, ok, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if !ok || err != nil {
			return
		}

		ringMembers := make([][]string, 1)
		if ringMembers[0], err = builder.CreateZetherRing(walletAddress.AddressEncoded, destinationAddress.EncodeAddr(), config_coins.NATIVE_ASSET_FULL, -1, -1); err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tx, err := builder.CreateZetherDelegateStakeTx(delegateAddress.PublicKey, delegatedStakingHasNewInfo, delegatePrivateKey, delegatedStakingNewPublicKey, delegatedStakingNewFee, []string{walletAddress.AddressEncoded}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{delegateAmount}, ringMembers, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		return
	}

	//cliPrivateClaim := func(cmd string) (err error) {
	//
	//	builder.showWarningIfNotSyncCLI()
	//
	//	walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Claim")
	//	if err != nil {
	//		return
	//	}
	//
	//	amount, ok, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Amount to Claim")
	//	if !ok || err != nil {
	//		return
	//	}
	//
	//	destinationAddress, ok := gui.GUI.OutputReadAddress("Destination Address")
	//	if !ok {
	//		return
	//	}
	//
	//	data, ok := builder.readData()
	//	if !ok {
	//		return
	//	}
	//
	//	propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
	//	if !ok {
	//		return
	//	}
	//
	//	fee, ok, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
	//	if !ok || err != nil {
	//		return
	//	}
	//
	//	ringMembers := make([][]string, 1)
	//	if ringMembers[0], err = builder.CreateZetherRing(walletAddress.AddressEncoded, destinationAddress.EncodeAddr(), assetId, -1, -1); err != nil {
	//		return
	//	}
	//
	//	ctx, cancel := context.WithCancel(context.Background())
	//	defer cancel()
	//
	//	tx, err := builder.CreateZetherClaimStakeTx( []byte{walletAddress.PrivateKey.Key}, [][]byte{assetId}, []uint64{amount}, []string{destinationAddress.EncodeAddr()}, []uint64{0}, ringMembers, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, propagate, true, true, false, ctx, func(status string) {
	//		gui.GUI.OutputWrite(status)
	//	})
	//	if err != nil {
	//		return
	//	}
	//
	//	gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
	//	return
	//}

	cliUpdateDelegate := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Update Delegate")
		if err != nil {
			return
		}

		nonce, ok := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", nil, true)
		if !ok {
			return
		}

		delegatedStakingHasNewInfo, ok := gui.GUI.OutputReadBool("New Delegate Info ? Type y/n")
		if !ok {
			return
		}

		var delegatedStakingNewPublicKey []byte
		var delegatedStakingNewFee uint64
		if delegatedStakingHasNewInfo {

			if delegatedStakingNewPublicKey, ok = gui.GUI.OutputReadBytes("Delegated Staking New PublicKey. Leave Empty to automatically derive pubKey", []int{20, 0}); !ok {
				return
			}

			if len(delegatedStakingNewPublicKey) == 0 {
				var derivedDelegatedStake *wallet_address.WalletAddressDelegatedStake
				if derivedDelegatedStake, err = builder.DeriveDelegatedStake(0, walletAddress.PublicKey); err != nil {
					return
				}
				delegatedStakingNewPublicKey = derivedDelegatedStake.PublicKey
			}

			delegatedStakingNewFee, ok = gui.GUI.OutputReadUint64("Delegated Staking New Fee. Leave empty for nothing", nil, true)
			if delegatedStakingNewFee > config_stake.DELEGATING_STAKING_FEES_MAX_VALUE {
				return errors.New("Invalid NewFee")
			}

		}

		delegatedStakingClaimAmount, ok, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Update Delegated Staking Amount")
		if !ok || err != nil {
			return
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		fee, ok, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if !ok || err != nil {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateUpdateDelegateTx(walletAddress.AddressEncoded, nonce, delegatedStakingClaimAmount, delegatedStakingHasNewInfo, delegatedStakingNewPublicKey, delegatedStakingNewFee, data, fee, propagate, true, true, false, func(status string) {
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

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, ok, err := builder.readAmount(config_coins.NATIVE_ASSET_FULL, "Amount")
		if !ok || err != nil {
			return
		}

		nonce, ok := gui.GUI.OutputReadUint64("Nonce. Leave for automatically detection", nil, true)
		if !ok {
			return
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		fee, ok, err := builder.readFees(config_coins.NATIVE_ASSET_FULL)
		if !ok || err != nil {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateUnstakeTx(walletAddress.AddressEncoded, nonce, amount, data, fee, propagate, true, true, false, func(status string) {
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
	gui.GUI.CommandDefineCallback("Update Delegate", cliUpdateDelegate, true)
	gui.GUI.CommandDefineCallback("Unstake", cliUnstake, true)

}
