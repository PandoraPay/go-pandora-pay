package routes

import (
	"context"
	"pandora-pay/builds/builds_data"
	"pandora-pay/builds/electron_helper/server/global"
)

func RouteWalletDecryptBalance(req *builds_data.WalletDecryptBalanceReq) (any, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	value, err := global.AddressBalanceDecryptor.DecryptBalance("wallet", req.PublicKey, req.PrivateKey, req.Balance, req.Asset, true, req.PreviousValue, true, ctx, func(status string) {})
	if err != nil {
		return nil, err
	}

	return value, nil
}
