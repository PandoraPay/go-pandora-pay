package balance_decryptor

import (
	"context"
	"errors"
	"math/big"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"runtime"
	"sync"
)

// table size cannot be more than 1<<24

type BalanceDecryptorType struct {
	once            *sync.Once
	tableComputedCn chan *LookupTable
	tableSize       int
	tableLookup     *LookupTable
	readyCn         chan struct{}
}

func (this *BalanceDecryptorType) TryDecryptBalance(p *bn256.G1, matchBalance uint64) bool {
	var acc bn256.G1
	acc.ScalarMult(crypto.G, new(big.Int).SetUint64(matchBalance))
	return acc.String() == p.String()
}

func (this *BalanceDecryptorType) DecryptBalance(p *bn256.G1, tryPreviousValue bool, previousBalance uint64, ctx context.Context, statusCallback func(string)) (uint64, error) {

	if tryPreviousValue {
		if this.TryDecryptBalance(p, previousBalance) {
			return previousBalance, nil
		}
	}

	tableLookup := this.SetTableSize(0, context.Background(), statusCallback)
	if tableLookup == nil {
		return 0, errors.New("It was stopped")
	}

	return tableLookup.Lookup(p, ctx, statusCallback)
}

func (this *BalanceDecryptorType) SetTableSize(newTableSize int, ctx context.Context, statusCallback func(string)) *LookupTable {

	this.once.Do(func() {

		if newTableSize == 0 {
			if runtime.GOARCH != "wasm" {
				newTableSize = 1 << 16 //4mb ram
			} else {
				newTableSize = 1 << 20 //32mb ram
			}
		}
		if newTableSize > 1<<24 {
			panic("Table Size is incorrect")
		}

		this.tableSize = newTableSize

		go func() {
			createLookupTable(1, newTableSize, this.tableComputedCn, ctx, statusCallback)
		}()

		tableLookup := <-this.tableComputedCn
		this.tableLookup = tableLookup
		close(this.readyCn)
		close(this.tableComputedCn)

	})

	<-this.readyCn

	return this.tableLookup
}

var BalanceDecryptor *BalanceDecryptorType

func init() {

	BalanceDecryptor = &BalanceDecryptorType{
		&sync.Once{},
		make(chan *LookupTable),
		0,
		nil,
		make(chan struct{}),
	}

}
