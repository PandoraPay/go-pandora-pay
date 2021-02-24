package genesis

import (
	"encoding/hex"
	"errors"
	"pandora-pay/block"
	"pandora-pay/config"
	"time"
)

type Genesis struct {
	PrevHash       string
	PrevKernelHash string
	Timestamp      uint64

	PublicKeyHash string
}

var genesisMainet = Genesis{
	PrevHash:       "e6849c309a8e48dd1518ce1f756b9feb0ce1be585510a32b40bcd6bec066d808",
	PrevKernelHash: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:      uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
}

var genesisTestnet = Genesis{
	PrevHash:       "f4a2f9d1a71d1dfc448be029e381df81acc2e80ebf3607e51c60f085b16ca34b",
	PrevKernelHash: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:      uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
}

var genesisDevnet = Genesis{
	PrevHash:       "cc423820a65ec26892c0a0c7f1a6e7731fb3ac76b9ad98ec775dd33c7271b443",
	PrevKernelHash: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:      uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
}

func CreateGenesisBlock() (*block.Block, error) {

	var genesis Genesis

	switch config.NETWORK_SELECTED {
	case config.MAIN_NET_NETWORK_BYTE:
		genesis = genesisMainet
	case config.TEST_NET_NETWORK_BYTE:
		genesis = genesisTestnet
	case config.DEV_NET_NETWORK_BYTE:
		genesis = genesisDevnet
	default:
		return nil, errors.New("Invalid network")
	}

	var block block.Block

	block.BlockHeader.Height = 0
	block.BlockHeader.Timestamp = genesis.Timestamp

	var buf []byte
	buf, _ = hex.DecodeString(genesis.PrevHash)
	copy(block.PrevHash[:], buf)

	buf, _ = hex.DecodeString(genesis.PrevKernelHash)
	copy(block.PrevKernelHash[:], buf)

	return &block, nil
}
