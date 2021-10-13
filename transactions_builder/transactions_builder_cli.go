package transactions_builder

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/wallet/wallet_address"
)

func (builder *TransactionsBuilder) showWarningIfNotSyncCLI() {
	if builder.chain.Sync.GetSyncTime() == 0 {
		gui.GUI.OutputWrite("Your node is not Sync yet. Wait for it to get sync.")
	}
}

func (builder *TransactionsBuilder) readFees() (out *TransactionsBuilderFeeFloat, ok bool) {

	fee := &TransactionsBuilderFeeFloat{}

	if fee.PerByteAuto, ok = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte. Type y\n"); !ok {
		return
	}
	if !fee.PerByteAuto {
		if fee.PerByte, ok = gui.GUI.OutputReadFloat64("Fee per byte", nil); !ok {
			return
		}

		if fee.PerByte == 0 {
			if fee.Fixed, ok = gui.GUI.OutputReadFloat64("Fee per byte", nil); !ok {
				return
			}
		}
	}

	return fee, true
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

func (builder *TransactionsBuilder) initCLI() {

	cliTransfer := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Transfer")
		if err != nil {
			return
		}

		assetId, ok := gui.GUI.OutputReadBytes("Asset. Leave empty for Native Asset", []int{0, config_coins.ASSET_LENGTH})
		if !ok {
			return
		}
		if len(assetId) != 0 && len(assetId) != 40 {
			return errors.New("Invalid AssetId")
		}

		amount, ok := gui.GUI.OutputReadFloat64("Amount", nil)
		if !ok {
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

		fee, ok := builder.readFees()
		if !ok {
			return
		}

		ringMembers := make([][]string, 1)
		if ringMembers[0], err = builder.CreateZetherRing(walletAddress.AddressEncoded, destinationAddress.EncodeAddr(), assetId, -1, -1); err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tx, err := builder.CreateZetherTx_Float([]string{walletAddress.AddressEncoded}, [][]byte{assetId}, []float64{amount}, []string{destinationAddress.EncodeAddr()}, []float64{0}, ringMembers, []*wizard.TransactionsWizardData{data}, []*TransactionsBuilderFeeFloat{fee}, propagate, true, true, false, ctx, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		if propagate {
			gui.GUI.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	cliDelegate := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
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

		delegateStakingUpdateAmount, ok := gui.GUI.OutputReadFloat64("Update Delegated Staking Amount", nil)
		if !ok {
			return
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		fee, ok := builder.readFees()
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateUpdateDelegateTx_Float(walletAddress.AddressEncoded, nonce, delegatedStakingNewPublicKey, delegatedStakingNewFee, delegateStakingUpdateAmount, data, fee, propagate, true, true, false, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		if propagate {
			gui.GUI.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	cliUnstake := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, ok := gui.GUI.OutputReadFloat64("Amount", nil)
		if !ok {
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

		fee, ok := builder.readFees()
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateUnstakeTx_Float(walletAddress.AddressEncoded, nonce, amount, data, fee, propagate, true, true, false, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		if propagate {
			gui.GUI.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	gui.GUI.CommandDefineCallback("Transfer", cliTransfer, true)
	gui.GUI.CommandDefineCallback("Delegate", cliDelegate, true)
	gui.GUI.CommandDefineCallback("Unstake", cliUnstake, true)

}
