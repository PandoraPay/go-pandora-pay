package wallet

import (
	"encoding/binary"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_forging"
	"pandora-pay/gui"
	"pandora-pay/recovery"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
	"time"
)

func (wallet *Wallet) processRefreshWallets() {

	recovery.SafeGo(func() {
		var err error

		for {

			if config_forging.FORGING_ENABLED {

				accsList := []*account.Account{}
				addressesList := []*wallet_address.WalletAddress{}
				var chainHeight uint64

				if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

					chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))

					dataStorage := data_storage.NewDataStorage(reader)

					var accs *accounts.Accounts
					if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
						return
					}

					visited := make(map[string]bool)
					for i := 0; i < 50; i++ {
						addr := wallet.GetRandomAddress()
						if visited[string(addr.PublicKey)] {
							continue
						}
						visited[string(addr.PublicKey)] = true

						var acc *account.Account
						if acc, err = accs.GetAccount(addr.PublicKey, chainHeight); err != nil {
							return
						}

						accsList = append(accsList, acc)
						addressesList = append(addressesList, addr)
					}

					return
				}); err != nil {
					gui.GUI.Error("Error processRefreshWallets", err)
				}

				for i, acc := range accsList {
					if err = wallet.refreshWalletAccount(acc, chainHeight, wallet.GetWalletAddressByPublicKey(addressesList[i].PublicKey, true)); err != nil {
						return
					}
				}

			}

			time.Sleep(2 * time.Minute)

		}
	})

}
