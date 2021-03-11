package genesis

import (
	"encoding/hex"
	"errors"
	"pandora-pay/blockchain/block"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"time"
)

type GenesisDataType struct {
	Hash          cryptography.Hash
	HashHex       string
	KernelHash    cryptography.Hash
	KernelHashHex string
	Timestamp     uint64
	Target        cryptography.Hash
	TargetHex     string
}

var genesisMainet = GenesisDataType{
	HashHex:       "e6849c309a8e48dd1518ce1f756b9feb0ce1be585510a32b40bcd6bec066d808",
	KernelHashHex: "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	TargetHex:     "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
}

var genesisTestnet = GenesisDataType{
	HashHex:       "f4a2f9d1a71d1dfc448be029e381df81acc2e80ebf3607e51c60f085b16ca34b",
	KernelHashHex: "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	TargetHex:     "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
}

var genesisDevnet = GenesisDataType{
	HashHex:       "cc423820a65ec26892c0a0c7f1a6e7731fb3ac76b9ad98ec775dd33c7271b443",
	KernelHashHex: "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	TargetHex:     "0FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
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
		MerkleHash:     cryptography.SHA3Hash([]byte{}),
		Timestamp:      GenesisData.Timestamp,
		PrevHash:       GenesisData.Hash,
		PrevKernelHash: GenesisData.KernelHash,
	}

	return &blk, nil
}

func GenesisInit() {

	var err error
	if GenesisData, err = getGenesis(); err != nil {
		gui.Fatal("Invalid Network for Genesis")
	}

	if globals.Arguments["--new-devnet"] == true {
		GenesisData.HashHex = hex.EncodeToString(helpers.RandomBytes(cryptography.HashSize))
		GenesisData.Timestamp = uint64(time.Now().Unix()) //the reason is to forge first block fast in tests
	}

	var buf []byte
	buf, _ = hex.DecodeString(GenesisData.HashHex)
	GenesisData.Hash = *cryptography.ConvertHash(buf)

	buf, _ = hex.DecodeString(GenesisData.KernelHashHex)
	GenesisData.KernelHash = *cryptography.ConvertHash(buf)

	buf, _ = hex.DecodeString(GenesisData.TargetHex)
	GenesisData.Target = *cryptography.ConvertHash(buf)

	if Genesis, err = CreateNewGenesisBlock(); err != nil {
		gui.Fatal("Error generating init Genesis")
	}
}
