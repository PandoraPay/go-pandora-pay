//go:build wasm
// +build wasm

package api_faucet

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/txs_builder"
	"pandora-pay/wallet"
)

type Faucet struct {
}

func NewFaucet(mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet, txsBuilder *txs_builder.TxsBuilder) (*Faucet, error) {
	return &Faucet{}, nil
}
