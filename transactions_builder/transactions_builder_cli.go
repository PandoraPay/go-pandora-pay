package transactions_builder

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
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

	//cliPrivateDelegate := func(cmd string) (err error) {
	//
	//	builder.showWarningIfNotSyncCLI()
	//
	//	walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Transfer")
	//	if err != nil {
	//		return
	//	}
	//
	//	assetId, ok := gui.GUI.OutputReadBytes("Asset. Leave empty for Native Asset", []int{0, config_coins.ASSET_LENGTH})
	//	if !ok {
	//		return
	//	}
	//	if len(assetId) != 0 && len(assetId) != 40 {
	//		return errors.New("Invalid AssetId")
	//	}
	//
	//	amount, ok := gui.GUI.OutputReadFloat64("Amount", nil)
	//	if !ok {
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
	//	fee, ok := builder.readFees()
	//	if !ok {
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
	//	tx, err := builder.CreateZetherClaimStakeTx_Float([]string{walletAddress.AddressEncoded}, [][]byte{assetId}, []float64{amount}, []string{destinationAddress.EncodeAddr()}, []float64{0}, ringMembers, []*wizard.TransactionsWizardData{data}, []*TransactionsBuilderFeeFloat{fee}, propagate, true, true, false, ctx, func(status string) {
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

		delegatedStakingNewPublicKey, ok := gui.GUI.OutputReadBytes("Delegate New Public Key. Use empty for not changing. Use '01' for generating a new one. ", []int{0, 1, cryptography.PublicKeySize})
		if !ok {
			return
		}

		if len(delegatedStakingNewPublicKey) == 1 {
			if bytes.Equal(delegatedStakingNewPublicKey, []byte{1}) {
				var derivedDelegatedStake *wallet_address.WalletAddressDelegatedStake
				if derivedDelegatedStake, err = builder.DeriveDelegatedStake(nonce, walletAddress.PublicKey); err != nil {
					return
				}
				delegatedStakingNewPublicKey = derivedDelegatedStake.PublicKey
			} else {
				return errors.New("Invalid value for New Public key Hash")
			}
		}

		var delegatedStakingNewFee uint64
		if len(delegatedStakingNewPublicKey) > 0 {
			number, ok := gui.GUI.OutputReadUint64("New Fee", nil, true)
			if !ok {
				return
			}
			delegatedStakingNewFee = number
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

		tx, err := builder.CreateUpdateDelegateTx(walletAddress.AddressEncoded, nonce, delegatedStakingNewPublicKey, delegatedStakingNewFee, delegatedStakingClaimAmount, data, fee, propagate, true, true, false, func(status string) {
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
	//gui.GUI.CommandDefineCallback("Private Delegate", cliPrivateDelegate, true)
	gui.GUI.CommandDefineCallback("Update Delegate", cliUpdateDelegate, true)
	gui.GUI.CommandDefineCallback("Unstake", cliUnstake, true)

}
