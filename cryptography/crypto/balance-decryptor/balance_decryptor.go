package balance_decryptor

import (
	"context"
	"errors"
	"github.com/tevino/abool"
	"math"
	"math/big"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"runtime"
	"strconv"
)

// table size cannot be more than 1<<24

type BalanceDecoderInfo struct {
	tableSize       int
	tableLookup     *generics.Value[*LookupTable]
	tableComputedCn chan *LookupTable
	ctx             context.Context
	readyCn         chan struct{}
	hasError        *abool.AtomicBool
}

type BalanceDecryptorType struct {
	info *generics.Value[*BalanceDecoderInfo]
}

func (self *BalanceDecryptorType) TryDecryptBalance(p *bn256.G1, matchBalance uint64) bool {
	var acc bn256.G1
	acc.ScalarMult(crypto.G, new(big.Int).SetUint64(matchBalance))
	return acc.String() == p.String()
}

func (self *BalanceDecryptorType) DecryptBalance(p *bn256.G1, previousBalance uint64, ctx context.Context, statusCallback func(string)) (uint64, error) {

	if self.TryDecryptBalance(p, previousBalance) {
		return previousBalance, nil
	}

	tableLookup := self.SetTableSize(0, ctx, statusCallback)
	if tableLookup == nil {
		return 0, errors.New("It was stopped")
	}

	return tableLookup.Lookup(p, ctx, statusCallback)
}

func (self *BalanceDecryptorType) SetTableSize(newTableSize int, ctx context.Context, statusCallback func(string)) *LookupTable {

	info := self.info.Load()
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
			&generics.Value[*LookupTable]{},
			make(chan *LookupTable),
			ctx,
			make(chan struct{}),
			abool.New(),
		}
		self.info.Store(info)

		gui.GUI.Info2Update("Decrypter", "Init... "+strconv.Itoa(int(math.Log2(float64(info.tableSize)))))

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
			info.tableLookup.Store(tableLookup)
			gui.GUI.Info2Update("Decrypter", "Ready "+strconv.Itoa(int(math.Log2(float64(info.tableSize)))))
			return tableLookup
		case <-ctx.Done():
			if info.hasError.SetToIf(false, true) {
				close(info.readyCn)
			}
			return nil
		}

	}

	select {
	case <-info.ctx.Done():
	}

	return info.tableLookup.Load()
}

var BalanceDecryptor *BalanceDecryptorType

func init() {

	BalanceDecryptor = &BalanceDecryptorType{
		&generics.Value[*BalanceDecoderInfo]{},
	}

	info := &BalanceDecoderInfo{
		tableSize: 0,
		readyCn:   make(chan struct{}),
		hasError:  abool.New(),
	}
	info.hasError.SetTo(true)
	BalanceDecryptor.info.Store(info)
}
