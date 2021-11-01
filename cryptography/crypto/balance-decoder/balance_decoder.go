package balance_decoder

import (
	"context"
	"errors"
	"github.com/tevino/abool"
	"math"
	"math/big"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"runtime"
	"strconv"
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

func (self *BalanceDecoderType) BalanceDecode(p *bn256.G1, previousBalance uint64, ctx context.Context, statusCallback func(string)) (uint64, error) {

	var acc bn256.G1
	acc.ScalarMult(crypto.G, new(big.Int).SetUint64(previousBalance))
	if acc.String() == p.String() {
		return previousBalance, nil
	}

	tableLookup := self.SetTableSize(0, ctx, statusCallback)
	if tableLookup == nil {
		return 0, errors.New("It was stopped")
	}

	return tableLookup.Lookup(p, ctx, statusCallback)
}

func (self *BalanceDecoderType) SetTableSize(newTableSize int, ctx context.Context, statusCallback func(string)) *LookupTable {

	info := self.info.Load().(*BalanceDecoderInfo)
	if info.tableSize == 0 || info.tableSize < newTableSize || info.hasError.IsSet() {

		if newTableSize == 0 {
			if runtime.GOARCH != "wasm" {
				newTableSize = 1 << 23 //32mb ram
			} else {
				newTableSize = 1 << 16 //4mb ram
			}
		}
		if newTableSize > 1<<24 {
			panic("Table Size is incorrect")
		}

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

		gui.GUI.Info2Update("Decoder", "Init... "+strconv.Itoa(int(math.Log2(float64(info.tableSize)))))

		if oldInfo != nil && oldInfo.hasError.SetToIf(false, true) {
			close(oldInfo.readyCn)
		}

		go func() {
			createLookupTable(1, newTableSize, info.tableComputedCn, info.readyCn, ctx, statusCallback)
		}()

		select {
		case tableLookup := <-info.tableComputedCn:
			if tableLookup == nil && info.hasError.SetToIf(false, true) {
				close(info.readyCn)
			}
			info.tableLookup = tableLookup
			gui.GUI.Info2Update("Decoder", "Ready "+strconv.Itoa(int(math.Log2(float64(info.tableSize)))))
			return tableLookup
		case <-ctx.Done():
			if info.hasError.SetToIf(false, true) {
				close(info.readyCn)
			}
			return nil
		}

	}
	return info.tableLookup
}

func CreateBalanceDecoder() *BalanceDecoderType {
	out := &BalanceDecoderType{
		&atomic.Value{},
	}
	info := &BalanceDecoderInfo{
		tableSize: 0,
		readyCn:   make(chan struct{}),
		hasError:  abool.New(),
	}
	info.hasError.SetTo(true)
	out.info.Store(info)

	return out
}

var BalanceDecoder *BalanceDecoderType

func init() {
	BalanceDecoder = CreateBalanceDecoder()
}
