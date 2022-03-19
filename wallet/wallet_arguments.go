package wallet

import (
	"errors"
	"pandora-pay/config/globals"
	"pandora-pay/wallet/wallet_address"
	"strconv"
	"strings"
)

func (wallet *Wallet) ProcessWalletArguments() (err error) {

	if str := globals.Arguments["--wallet-encrypt"]; str != nil {
		v := strings.Split(str.(string), ",")

		var diff int
		if diff, err = strconv.Atoi(v[1]); err != nil {
			return
		}

		if err = wallet.Encryption.Encrypt(v[0], diff); err != nil {
			return
		}
	}

	if password := globals.Arguments["--wallet-decrypt"]; password != nil {
		if err = wallet.loadWallet(password.(string), true); err != nil {
			return
		}
	}

	if globals.Arguments["--wallet-remove-encryption"] == true {
		if err = wallet.Encryption.RemoveEncryption(); err != nil {
			return
		}
	}

	if str := globals.Arguments["--wallet-export-shared-staked-address"]; str != nil {
		v := strings.Split(str.(string), ",")

		var addr *wallet_address.WalletAddress

		if v[0] == "auto" {
			if addr, err = wallet.GetFirstStakedAddress(true); err != nil {
				return
			}
		} else {
			var index int
			if index, err = strconv.Atoi(v[0]); err != nil {
				return
			} else {
				if addr, err = wallet.GetWalletAddress(index, true); err != nil {
					return
				}
			}
			if addr == nil {
				if addr, err = wallet.GetWalletAddressByEncodedAddress(v[0], true); err != nil {
					return
				}
			}
		}

		if addr == nil {
			return errors.New("Address specified by --wallet-export-shared-staked-address was not found")
		}
		if err = wallet.exportSharedStakedAddress(addr, v[2], false); err != nil {
			return
		}

	}

	return
}
