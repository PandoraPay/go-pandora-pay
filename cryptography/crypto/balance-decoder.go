package crypto

import (
	"context"
	"errors"
	"github.com/tevino/abool"
	"math/big"
	"pandora-pay/cryptography/bn256"
	"runtime"
	"sync/atomic"
)

// table size cannot be more than 1<<24

type BalanceDecoderInfo struct {
	tableSize       int
	tableLookup     *LookupTable
	tableComputedCn chan *LookupTable
	ctx             context.Context
	readyCn         chan struct{}
	hasError        *abool.AtomicBool
}

type BalanceDecoderType struct {
	info *atomic.Value //*BalanceDecoderInfo
}

func (self *BalanceDecoderType) BalanceDecode(p *bn256.G1, previousBalance uint64, ctx context.Context) (uint64, error) {

	var acc bn256.G1
	acc.ScalarMult(G, new(big.Int).SetUint64(previousBalance))
	if acc.String() == p.String() {
		return previousBalance, nil
	}

	info := self.info.Load().(*BalanceDecoderInfo)
	if info.tableSize == 0 || info.hasError.IsSet() {
		if err := self.SetTableSize(0, ctx); err != nil {
			return 0, err
		}
		info = self.info.Load().(*BalanceDecoderInfo)
	}
	select {
	case <-info.readyCn:
		if info.hasError.IsSet() {
			return 0, errors.New("Suspended")
		}
		return info.tableLookup.Lookup(p, ctx)
	case <-ctx.Done():
		return 0, errors.New("Suspended")
	}

}

func (self *BalanceDecoderType) SetTableSize(newTableSize int, ctx context.Context) error {

	if newTableSize == 0 {
		if runtime.GOARCH != "wasm" {
			newTableSize = 1 << 22 //32mb ram
		} else {
			newTableSize = 1 << 19 //4mb ram
		}
	}
	if newTableSize > 1<<24 {
		return errors.New("Table Size is incorrect")
	}

	info := self.info.Load().(*BalanceDecoderInfo)
	if info.tableSize == 0 || info.tableSize != newTableSize || info.hasError.IsSet() {

		oldInfo := info

		info = &BalanceDecoderInfo{
			newTableSize,
			nil,
			make(chan *LookupTable),
			ctx,
			make(chan struct{}),
			abool.New(),
		}
		self.info.Store(info)

		if oldInfo != nil && oldInfo.hasError.SetToIf(false, true) {
			close(oldInfo.readyCn)
		}

		createLookupTable(1, newTableSize, info.tableComputedCn, info.readyCn, ctx)

		select {
		case tableLookup := <-info.tableComputedCn:
			info.tableLookup = tableLookup
		case <-ctx.Done():
			if info.hasError.SetToIf(false, true) {
				close(info.readyCn)
			}
			return errors.New("it was stopped")
		case <-info.readyCn:
			return nil
		}

	}

	return nil
}

func CreateBalanceDecoder() *BalanceDecoderType {
	out := &BalanceDecoderType{
		&atomic.Value{},
	}
	out.info.Store(&BalanceDecoderInfo{
		tableSize: 0,
		readyCn:   make(chan struct{}),
		hasError:  abool.New(),
	})
	return out
}

var BalanceDecoder *BalanceDecoderType

func init() {
	BalanceDecoder = CreateBalanceDecoder()
}
