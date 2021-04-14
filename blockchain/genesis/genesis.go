package genesis

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"pandora-pay/blockchain/block"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/config/stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/wallet"
	"strconv"
	"time"
)

type GenesisDataAirDropType struct {
	PublicKeyHash               []byte //20 byte
	Amount                      uint64
	DelegatedStakePublicKeyHash []byte
}

type GenesisDataType struct {
	Hash          []byte //32 byte
	HashHex       string
	KernelHash    []byte //32 byte
	KernelHashHex string
	Timestamp     uint64
	Target        []byte //32 byte
	TargetHex     string
	AidDrops      []*GenesisDataAirDropType
}

var genesisMainet = GenesisDataType{
	HashHex:       "e6849c309a8e48dd1518ce1f756b9feb0ce1be585510a32b40bcd6bec066d808",
	KernelHashHex: "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	TargetHex:     "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	AidDrops:      []*GenesisDataAirDropType{},
}

var genesisTestnet = GenesisDataType{
	HashHex:       "f4a2f9d1a71d1dfc448be029e381df81acc2e80ebf3607e51c60f085b16ca34b",
	KernelHashHex: "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	TargetHex:     "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	AidDrops:      []*GenesisDataAirDropType{},
}

var genesisDevnet = GenesisDataType{
	HashHex:       "cc423820a65ec26892c0a0c7f1a6e7731fb3ac76b9ad98ec775dd33c7271b443",
	KernelHashHex: "000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	Timestamp:     uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	TargetHex:     "0FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	AidDrops:      []*GenesisDataAirDropType{},
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

func GenesisInit(wallet *wallet.Wallet) (err error) {

	var data []byte

	if GenesisData, err = getGenesis(); err != nil {
		return
	}

	if globals.Arguments["--new-devnet"] == true {

		var file *os.File
		if _, err = os.Stat("./genesis.data"); os.IsNotExist(err) {

			HashHex := hex.EncodeToString(helpers.RandomBytes(cryptography.HashSize))
			Timestamp := uint64(time.Now().Unix()) //the reason is to forge first block fast in tests

			walletAddress, delegatedStakePublicKeyHash, err2 := wallet.GetFirstWalletForDevnetGenesisAirdrop()
			if err2 != nil {
				return err2
			}

			AidDrops := make([]*GenesisDataAirDropType, 0)

			amount := stake.GetRequiredStake(0)

			AidDrops = append(AidDrops, &GenesisDataAirDropType{
				PublicKeyHash:               walletAddress.Address.PublicKeyHash,
				Amount:                      amount,
				DelegatedStakePublicKeyHash: delegatedStakePublicKeyHash,
			})

			if file, err = os.OpenFile("./genesis.data", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
				return
			}

			if _, err = file.WriteString("1\n"); err != nil {
				return
			}
			if _, err = file.WriteString(HashHex + "\n"); err != nil {
				return
			}
			if _, err = file.WriteString(strconv.FormatUint(Timestamp, 10) + "\n"); err != nil {
				return
			}

			if data, err = json.Marshal(AidDrops); err != nil {
				return
			}

			if _, err = file.WriteString(hex.EncodeToString(data) + "\n"); err != nil {
				return
			}
			if err = file.Close(); err != nil {
				return
			}
		}

		if file, err = os.OpenFile("./genesis.data", os.O_RDONLY, 0666); err != nil {
			return
		}

		scanner := bufio.NewScanner(file)
		scanner.Scan()

		version := scanner.Text()
		if version != "1" {
			return errors.New("Genesis config data version is invalid")
		}
		scanner.Scan()

		GenesisData.HashHex = scanner.Text()
		scanner.Scan()

		if GenesisData.Timestamp, err = strconv.ParseUint(scanner.Text(), 10, 64); err != nil {
			return
		}
		scanner.Scan()

		data, err = hex.DecodeString(scanner.Text())
		if err = json.Unmarshal(data, &GenesisData.AidDrops); err != nil {
			return
		}
	}

	if GenesisData.Hash, err = hex.DecodeString(GenesisData.HashHex); err != nil {
		return
	}

	if GenesisData.KernelHash, err = hex.DecodeString(GenesisData.KernelHashHex); err != nil {
		return
	}

	if GenesisData.Target, err = hex.DecodeString(GenesisData.TargetHex); err != nil {
		return
	}

	if Genesis, err = CreateNewGenesisBlock(); err != nil {
		return
	}

	return
}
