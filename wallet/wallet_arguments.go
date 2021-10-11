package wallet

import (
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

	if str := globals.Arguments["--wallet-derive-delegated-stake"]; str != nil {
		v := strings.Split(str.(string), ",")

		var addr *wallet_address.WalletAddress

		var index int
		if index, err = strconv.Atoi(v[0]); err != nil {
			return
		} else {
			if addr, err = wallet.GetWalletAddress(index); err != nil {
				return
			}
		}

		if addr == nil {
			if addr, err = wallet.GetWalletAddressByEncodedAddress(v[0]); err != nil {
				return
			}
		}

		var nonce uint64
		if nonce, err = strconv.ParseUint(v[1], 10, 64); err != nil {
			return
		}

		if err = wallet.deriveDelegatedStake(addr, nonce, v[2], false); err != nil {
			return
		}

	}

	return
}
