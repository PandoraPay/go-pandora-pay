package genesis

import (
	"encoding/hex"
	"errors"
	"pandora-pay/block"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"time"
)

type GenesisType struct {
	Hash          crypto.Hash
	HashHex       string
	KernelHash    crypto.Hash
	KernelHashHex string
	Timestamp     uint64
	Difficulty    uint64
}

var genesisMainet = GenesisType{
	HashHex:       "e6849c309a8e48dd1518ce1f756b9feb0ce1be585510a32b40bcd6bec066d808",
	KernelHashHex: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
	Difficulty:    1,
}

var genesisTestnet = GenesisType{
	HashHex:       "f4a2f9d1a71d1dfc448be029e381df81acc2e80ebf3607e51c60f085b16ca34b",
	KernelHashHex: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
	Difficulty:    1,
}

var genesisDevnet = GenesisType{
	HashHex:       "cc423820a65ec26892c0a0c7f1a6e7731fb3ac76b9ad98ec775dd33c7271b443",
	KernelHashHex: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
	Difficulty:    1,
}

var Genesis *GenesisType

func getGenesis() (*GenesisType, error) {

	switch config.NETWORK_SELECTED {
	case config.MAIN_NET_NETWORK_BYTE:
		return &genesisMainet, nil
	case config.TEST_NET_NETWORK_BYTE:
		return &genesisTestnet, nil
	case config.DEV_NET_NETWORK_BYTE:
		return &genesisDevnet, nil
	default:
		return nil, errors.New("Invalid Network")
	}
}

func CreateGenesisBlock() (*block.Block, error) {

	var blockHeader = block.BlockHeader{
		Version: 0,
		Height:  0,
	}

	var blk = block.Block{
		BlockHeader:    blockHeader,
		MerkleHash:     crypto.SHA3Hash([]byte{}),
		Timestamp:      Genesis.Timestamp,
		PrevHash:       Genesis.Hash,
		PrevKernelHash: Genesis.KernelHash,
	}

	return &blk, nil
}

func GenesisInit() {

	var err error
	Genesis, err = getGenesis()
	if err != nil {
		gui.Fatal("Invalid Network for Genesis")
	}

	var buf []byte
	buf, _ = hex.DecodeString(Genesis.HashHex)
	copy(Genesis.Hash[:], buf)

	buf, _ = hex.DecodeString(Genesis.KernelHashHex)
	copy(Genesis.KernelHash[:], buf)

}
