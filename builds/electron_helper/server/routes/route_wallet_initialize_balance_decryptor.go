package routes

import (
	"context"
	"pandora-pay/builds/builds_data"
	"pandora-pay/cryptography/crypto/balance_decryptor"
)

func RouteWalletInitializeBalanceDecryptor(req *builds_data.WalletInitializeBalanceDecryptorReq) (any, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	balance_decryptor.BalanceDecryptor.SetTableSize(req.TableSize, ctx, func(status string) {})

	return true, nil
}
