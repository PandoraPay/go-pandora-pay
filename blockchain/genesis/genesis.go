package genesis

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/config/stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/wallet"
	"runtime"
	"strings"
	"time"
)

type GenesisDataAirDropType struct {
	PublicKeyHash               helpers.HexBytes `json:"publicKeyHash"` //20 byte
	Amount                      uint64           `json:"amount"`
	DelegatedStakePublicKeyHash helpers.HexBytes `json:"delegatedStakePublicKeyHash"`
}

type GenesisDataType struct {
	Hash       helpers.HexBytes          `json:"hash"`       //32 byte
	KernelHash helpers.HexBytes          `json:"kernelHash"` //32 byte
	Timestamp  uint64                    `json:"timestamp"`
	Target     helpers.HexBytes          `json:"target"` //32 byte
	AirDrops   []*GenesisDataAirDropType `json:"airDrops"`
}

var genesisMainet = GenesisDataType{
	Hash:       helpers.DecodeHex("e6849c309a8e48dd1518ce1f756b9feb0ce1be585510a32b40bcd6bec066d808"),
	KernelHash: helpers.DecodeHex("000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
	Timestamp:  uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	Target:     helpers.DecodeHex("000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
	AirDrops:   []*GenesisDataAirDropType{},
}

var genesisTestnet = GenesisDataType{
	Hash:       helpers.DecodeHex("f4a2f9d1a71d1dfc448be029e381df81acc2e80ebf3607e51c60f085b16ca34b"),
	KernelHash: helpers.DecodeHex("000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
	Timestamp:  uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	Target:     helpers.DecodeHex("000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
	AirDrops:   []*GenesisDataAirDropType{},
}

var genesisDevnet = GenesisDataType{
	Hash:       helpers.DecodeHex("cc423820a65ec26892c0a0c7f1a6e7731fb3ac76b9ad98ec775dd33c7271b443"),
	KernelHash: helpers.DecodeHex("000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
	Timestamp:  uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	Target:     helpers.DecodeHex("0FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
	AirDrops:   []*GenesisDataAirDropType{},
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

	var blk = block.Block{
		BlockHeader: &block.BlockHeader{
			Version: 0,
			Height:  0,
		},
		MerkleHash:     cryptography.SHA3Hash([]byte{}),
		Timestamp:      GenesisData.Timestamp,
		PrevHash:       GenesisData.Hash,
		PrevKernelHash: GenesisData.KernelHash,
	}

	return &blk, nil
}

func createNewGenesis(v []string) (err error) {

	var file *os.File
	if file, err = os.Create(v[0]); err != nil {
		return
	}

	defer file.Close()

	GenesisData.Hash = helpers.RandomBytes(cryptography.HashSize)
	GenesisData.Timestamp = uint64(time.Now().Unix()) //the reason is to forge first block fast in tests

	amount := 100 * stake.GetRequiredStake(0)
	for i := 1; i < len(v); i++ {

		if err = func() (err error) {

			var file2 *os.File
			if file2, err = os.OpenFile(v[i], os.O_RDONLY, 0666); err != nil {
				return
			}

			defer file2.Close()

			scanner := bufio.NewScanner(file2)
			scanner.Scan()

			delegatedStakeOutput := &wallet.DelegatedStakeOutput{}
			if err = json.Unmarshal(scanner.Bytes(), delegatedStakeOutput); err != nil {
				return
			}

			GenesisData.AirDrops = append(GenesisData.AirDrops, &GenesisDataAirDropType{
				PublicKeyHash:               delegatedStakeOutput.AddressPublicKeyHash,
				Amount:                      amount,
				DelegatedStakePublicKeyHash: delegatedStakeOutput.DelegatedStakePublicKeyHash,
			})

			return
		}(); err != nil {
			return
		}

	}

	var data []byte
	if data, err = json.Marshal(GenesisData); err != nil {
		return
	}

	if _, err = file.Write(data); err != nil {
		return
	}

	return
}

func createSimpleGenesis(wallet *wallet.Wallet) (err error) {

	var file *os.File

	if globals.Arguments["--new-devnet"] == true {
		return errors.New("Genesis Data was not found and --new-devnet is missing")
	}

	GenesisData.Hash = helpers.RandomBytes(cryptography.HashSize)
	GenesisData.Timestamp = uint64(time.Now().Unix()) //the reason is to forge first block fast in tests

	var walletPublicKeyHash, delegatedStakePublicKeyHash []byte
	if walletPublicKeyHash, delegatedStakePublicKeyHash, err = wallet.GetFirstWalletForDevnetGenesisAirdrop(); err != nil {
		return
	}

	amount := 100 * stake.GetRequiredStake(0)
	GenesisData.AirDrops = append(GenesisData.AirDrops, &GenesisDataAirDropType{
		PublicKeyHash:               walletPublicKeyHash,
		Amount:                      amount,
		DelegatedStakePublicKeyHash: delegatedStakePublicKeyHash,
	})

	if file, err = os.OpenFile("./genesis.data", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return
	}
	defer file.Close()

	var data []byte
	if data, err = json.Marshal(GenesisData); err != nil {
		return
	}

	if _, err = file.Write(data); err != nil {
		return
	}

	return
}

func GenesisInit(w *wallet.Wallet) (err error) {

	if GenesisData, err = getGenesis(); err != nil {
		return
	}

	if dataArguments := globals.Arguments["--create-new-genesis"]; dataArguments != nil {
		if err = createNewGenesis(strings.Split(dataArguments.(string), ",")); err != nil {
			return
		}
	}

	if dataArgument := globals.Arguments["--set-genesis"]; dataArgument != nil {

		data := []byte(dataArgument.(string))

		if string(data) == "file" && runtime.GOARCH != "wasm" {

			if _, err = os.Stat("./genesis.data"); os.IsNotExist(err) {
				if err = createSimpleGenesis(w); err != nil {
					return
				}
			}

			var file *os.File
			if file, err = os.OpenFile("./genesis.data", os.O_RDONLY, 0666); err != nil {
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			scanner.Scan()

			data = scanner.Bytes()
		}

		if err = json.Unmarshal(data, &GenesisData); err != nil {
			return
		}

	}

	if Genesis, err = CreateNewGenesisBlock(); err != nil {
		return
	}

	return
}
