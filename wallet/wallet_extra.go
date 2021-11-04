package wallet

import (
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/recovery"
)

func (wallet *Wallet) updateAccountsChanges() {

	//recovery.SafeGo(func() {
	//	var err error
	//
	//	updateAccountsCn := wallet.updateAccounts.AddListener()
	//	defer wallet.updateAccounts.RemoveChannel(updateAccountsCn)
	//
	//	for {
	//		accsCollectionData, ok := <-updateAccountsCn
	//		if !ok {
	//			return
	//		}
	//
	//		accsCollection := accsCollectionData.(*accounts.AccountsCollection)
	//
	//		wallet.Lock()
	//		accsMap := accsCollection.GetAllMaps()
	//		for _, accs := range accsMap {
	//			for k, v := range accs.HashMap.Committed {
	//				if wallet.addressesMap[k] != nil {
	//
	//					if v.Stored == "update" {
	//						acc := v.Element.(*account.Account)
	//						if err = wallet.refreshWalletAccount(acc, wallet.addressesMap[k], false); err != nil {
	//							return
	//						}
	//					} else if v.Stored == "delete" {
	//						if err = wallet.refreshWalletAccount(nil, wallet.addressesMap[k], false); err != nil {
	//							return
	//						}
	//					}
	//
	//				}
	//			}
	//		}
	//		wallet.Unlock()
	//	}
	//})

	recovery.SafeGo(func() {
		var err error

		updatePlainAccountsCn := wallet.updatePlainAccounts.AddListener()
		defer wallet.updatePlainAccounts.RemoveChannel(updatePlainAccountsCn)

		for {
			plainAccsData, ok := <-updatePlainAccountsCn
			if !ok {
				return
			}

			plainAccs := plainAccsData.(*plain_accounts.PlainAccounts)

			wallet.Lock()
			for k, v := range plainAccs.HashMap.Committed {
				if wallet.addressesMap[k] != nil {

					if v.Stored == "update" {
						acc := v.Element.(*plain_account.PlainAccount)
						if err = wallet.refreshWalletPlainAccount(acc, wallet.addressesMap[k], false); err != nil {
							return
						}
					} else if v.Stored == "delete" {
						if err = wallet.refreshWalletPlainAccount(nil, wallet.addressesMap[k], false); err != nil {
							return
						}
					}

				}
			}
			wallet.Unlock()
		}
	})

}
