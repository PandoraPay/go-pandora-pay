package cryptolib

import (
	"errors"
	"pandora-pay/cryptography/bn256"
	"runtime"
	"sync/atomic"
)

// table size cannot be more than 1<<24

type BalanceDecoderInfo struct {
	tableSize       int
	tableLookup     *LookupTable
	tableComputedCn chan *LookupTable
	stopCn          chan bool
	readyCn         chan struct{}
}

type BalanceDecoderType struct {
	info *atomic.Value //*BalanceDecoderInfo
}

func (self *BalanceDecoderType) BalanceDecode(p *bn256.G1, previousBalance uint64) uint64 {
	info := self.info.Load().(*BalanceDecoderInfo)
	if info.tableSize == 0 {
		if err := self.SetTableSize(0); err != nil {
			panic(err)
		}
		info = self.info.Load().(*BalanceDecoderInfo)
	}
	select {
	case <-info.readyCn:
	}
	return info.tableLookup.Lookup(p, previousBalance)
}

func (self *BalanceDecoderType) SetTableSize(newTableSize int) error {

	if newTableSize == 0 {
		if runtime.GOARCH != "wasm" {
			newTableSize = 1 << 22 //32mb ram
		} else {
			newTableSize = 1 << 20 //8mb ram
		}
	}
	if newTableSize > 1<<24 {
		return errors.New("Table Size is incorrect")
	}

	info := self.info.Load().(*BalanceDecoderInfo)
	if info.tableSize == 0 || info.tableSize != newTableSize {

		oldInfo := info

		info = &BalanceDecoderInfo{
			newTableSize,
			nil,
			make(chan *LookupTable),
			make(chan bool),
			make(chan struct{}),
		}
		self.info.Store(info)

		if oldInfo != nil && oldInfo.stopCn != nil {
			close(oldInfo.stopCn)
		}

		go func() {
			createLookupTable(1, newTableSize, info.tableComputedCn, info.stopCn, info.readyCn)
		}()

		select {
		case tableLookup := <-info.tableComputedCn:
			info.tableLookup = tableLookup
		case <-info.readyCn:
			if info.tableLookup == nil {
				return errors.New("it was stopped")
			}
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
	})
	return out
}

var BalanceDecoder *BalanceDecoderType

func init() {
	BalanceDecoder = CreateBalanceDecoder()
}
