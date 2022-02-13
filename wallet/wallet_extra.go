package wallet

import (
	"encoding/binary"
	"math/rand"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/gui"
	"pandora-pay/recovery"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"time"
)

func (wallet *Wallet) processRefreshWallets() {

	recovery.SafeGo(func() {
		var err error

		for {

			if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

				chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

				plainAccs := plain_accounts.NewPlainAccounts(reader)
				wallet.Lock.Lock()
				defer wallet.Lock.Unlock()

				visited := make(map[int]bool)
				for i := 0; i < 50; i++ {
					index := rand.Intn(len(wallet.Addresses))
					if visited[index] {
						continue
					}
					visited[index] = true

					addr := wallet.Addresses[index]

					var plainAcc *plain_account.PlainAccount
					if plainAcc, err = plainAccs.GetPlainAccount(addr.PublicKey, chainHeight); err != nil {
						return
					}

					if err = wallet.refreshWalletPlainAccount(plainAcc, addr, false); err != nil {
						return
					}
				}

				return
			}); err != nil {
				gui.GUI.Error("Error processRefreshWallets", err)
			}

			time.Sleep(2 * time.Minute)

		}
	})

}
