//go:build wasm
// +build wasm

package api_faucet

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

type Faucet struct {
}

func CreateFaucet(mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*Faucet, error) {
	return &Faucet{}, nil
}
