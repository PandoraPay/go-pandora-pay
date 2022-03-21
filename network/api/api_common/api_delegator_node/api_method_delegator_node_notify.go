package api_delegator_node

import (
	"encoding/binary"
	"errors"
	"net/http"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/config_stake"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
)

type ApiDelegatorNodeNotifyRequest struct {
	SharedStakedPrivateKey helpers.Base64 `json:"sharedStakedPrivateKey" msgpack:"sharedStakedPrivateKey"`
	SharedStakedBalance    uint64         `json:"sharedStakedBalance" msgpack:"sharedStakedBalance"`
}

type ApiDelegatorNodeNotifyReply struct {
	Result bool `json:"result" msgpack:"result"`
}

func (api *DelegatorNode) DelegatorNotify(r *http.Request, args *ApiDelegatorNodeNotifyRequest, reply *ApiDelegatorNodeNotifyReply, authenticated bool) (err error) {

	if config_nodes.DELEGATOR_REQUIRE_AUTH && !authenticated {
		return errors.New("Invalid User or Password")
	}

	sharedStakedPrivateKey := &addresses.PrivateKey{args.SharedStakedPrivateKey}
	sharedStakedPublicKey := sharedStakedPrivateKey.GeneratePublicKey()

	addr := api.wallet.GetWalletAddressByPublicKey(sharedStakedPublicKey, true)
	if addr != nil && addr.PrivateKey == nil {
		reply.Result = true
		return
	}

	var acc *account.Account
	var chainHeight uint64

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		dataStorage := data_storage.NewDataStorage(reader)

		var reg *registration.Registration
		if reg, err = dataStorage.Regs.GetRegistration(sharedStakedPublicKey); err != nil {
			return
		}
		if reg == nil {
			return errors.New("Registration doesn't exist")
		}

		if !reg.Staked {
			return errors.New("Account is not staked")
		}

		var accs *accounts.Accounts
		if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
			return
		}

		if acc, err = accs.GetAccount(sharedStakedPublicKey); err != nil {
			return
		}
		if acc == nil {
			return errors.New("Account doesn't exist")
		}

		return nil

	}); err != nil {
		return err
	}

	if !sharedStakedPrivateKey.TryDecryptBalance(acc.Balance.Amount, args.SharedStakedBalance) {
		return errors.New("Decrypt Balance Doesn't match. Try again")
	}

	if args.SharedStakedBalance < config_stake.GetRequiredStake(chainHeight) {
		return errors.New("Your stake is not accepted because you will need at least the minimum staking amount")
	}

	if err = api.wallet.AddSharedStakedAddress(&wallet_address.WalletAddress{
		wallet_address.VERSION_NORMAL,
		"Delegated Stake",
		0,
		false,
		nil,
		sharedStakedPrivateKey,
		nil,
		nil,
		sharedStakedPublicKey,
		true,
		false,
		nil,
		&wallet_address.WalletAddressSharedStaked{
			sharedStakedPrivateKey,
			sharedStakedPublicKey,
		},
		"",
		"",
	}, true); err != nil {
		return
	}

	reply.Result = true

	return nil
}
