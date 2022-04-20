package wallet

import (
	"encoding/binary"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
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

				plainAccsList := []*plain_account.PlainAccount{}
				addressesList := []*wallet_address.WalletAddress{}
				var chainHeight uint64

				if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

					chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))

					dataStorage := data_storage.NewDataStorage(reader)

					visited := make(map[string]bool)
					for i := 0; i < 50; i++ {
						addr := wallet.GetRandomAddress()
						if visited[string(addr.PublicKey)] {
							continue
						}
						visited[string(addr.PublicKeyHash)] = true

						var plainAcc *plain_account.PlainAccount

						if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(addr.PublicKeyHash); err != nil {
							return
						}

						plainAccsList = append(plainAccsList, plainAcc)
						addressesList = append(addressesList, addr)
					}

					return
				}); err != nil {
					gui.GUI.Error("Error processRefreshWallets", err)
				}

				for i, plainAcc := range plainAccsList {
					if err = wallet.refreshWalletAccount(plainAcc, chainHeight, addressesList[i]); err != nil {
						return
					}
				}

			}

			time.Sleep(2 * time.Minute)

		}
	})

}
