// +build wasm

package forging

import (
	"github.com/tevino/abool"
	"math/big"
	"pandora-pay/blockchain/accounts"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/mempool"
)

type Forging struct {
	mempool    *mempool.Mempool
	Wallet     *ForgingWallet
	started    *abool.AtomicBool
	workCn     chan *ForgingWork
	SolutionCn chan *block_complete.BlockComplete
}

type ForgingWallet struct {
}

type ForgingWork struct {
	blkComplete *block_complete.BlockComplete
	target      *big.Int
}

func ForgingInit(mempool *mempool.Mempool) (forging *Forging, err error) {
	return &Forging{
		Wallet: &ForgingWallet{},
	}, nil
}

func (forging *Forging) StartForging() bool {
	return true
}

func (forging *Forging) StopForging() bool {
	return true
}

func (forging *Forging) ForgingNewWork(blkComplete *block_complete.BlockComplete, target *big.Int) {
}

func (forging *Forging) Close() {
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKeyHash []byte) {
}

func (w *ForgingWallet) RemoveWallet(delegatedPublicKeyHash []byte) { //20 byte
}

func (w *ForgingWallet) UpdateAccountsChanges(accs *accounts.Accounts) (err error) {
	return
}

func (w *ForgingWallet) ProcessUpdates() (err error) {
	return
}
