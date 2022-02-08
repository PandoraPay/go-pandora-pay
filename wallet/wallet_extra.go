package wallet

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/recovery"
)

func (wallet *Wallet) updateAccountsChanges() {

	recovery.SafeGo(func() {
		var err error

		updatePlainAccountsCn := wallet.updatePlainAccounts.AddListener()
		defer wallet.updatePlainAccounts.RemoveChannel(updatePlainAccountsCn)

		for {
			plainAccs, ok := <-updatePlainAccountsCn
			if !ok {
				return
			}

			for k, v := range plainAccs.HashMap.Committed {
				if wallet.GetWalletAddressByPublicKey([]byte(k), true) != nil {

					if v.Stored == "update" {
						acc := v.Element.(*plain_account.PlainAccount)
						if err = wallet.refreshWalletPlainAccount(acc, wallet.addressesMap[k], true); err != nil {
							return
						}
					} else if v.Stored == "delete" {
						if err = wallet.refreshWalletPlainAccount(nil, wallet.addressesMap[k], true); err != nil {
							return
						}
					}

				}
			}

		}
	})

}
