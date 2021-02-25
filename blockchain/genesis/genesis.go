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

type GenesisDataType struct {
	Hash          crypto.Hash
	HashHex       string
	KernelHash    crypto.Hash
	KernelHashHex string
	Timestamp     uint64
	Target        crypto.Hash
	Difficulty    uint64
}

var genesisMainet = GenesisDataType{
	HashHex:       "e6849c309a8e48dd1518ce1f756b9feb0ce1be585510a32b40bcd6bec066d808",
	KernelHashHex: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Difficulty:    1,
	Timestamp:     uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
}

var genesisTestnet = GenesisDataType{
	HashHex:       "f4a2f9d1a71d1dfc448be029e381df81acc2e80ebf3607e51c60f085b16ca34b",
	KernelHashHex: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Difficulty:    1,
	Timestamp:     uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
}

var genesisDevnet = GenesisDataType{
	HashHex:       "cc423820a65ec26892c0a0c7f1a6e7731fb3ac76b9ad98ec775dd33c7271b443",
	KernelHashHex: "0000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Difficulty:    1,
	Timestamp:     uint64(time.Date(2021, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
}

var GenesisData *GenesisDataType
var Genesis *block.Block

func getGenesis() (*GenesisDataType, error) {

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

func CreateNewGenesisBlock() (*block.Block, error) {

	var blockHeader = block.BlockHeader{
		Version: 0,
		Height:  0,
	}

	var blk = block.Block{
		BlockHeader:    blockHeader,
		MerkleHash:     crypto.SHA3Hash([]byte{}),
		Timestamp:      GenesisData.Timestamp,
		PrevHash:       GenesisData.Hash,
		PrevKernelHash: GenesisData.KernelHash,
	}

	return &blk, nil
}

func GenesisInit() {

	var err error
	GenesisData, err = getGenesis()
	if err != nil {
		gui.Fatal("Invalid Network for Genesis")
	}

	var buf []byte
	buf, _ = hex.DecodeString(GenesisData.HashHex)
	copy(GenesisData.Hash[:], buf)

	buf, _ = hex.DecodeString(GenesisData.KernelHashHex)
	copy(GenesisData.KernelHash[:], buf)

	Genesis, err = CreateNewGenesisBlock()
	if err != nil {
		gui.Fatal("Error generating init Genesis")
	}
}
