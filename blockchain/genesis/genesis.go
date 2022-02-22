package genesis

import (
	"encoding/json"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"io/ioutil"
	"os"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/config"
	"pandora-pay/config/config_stake"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"runtime"
	"strings"
	"time"
)

type DelegatedStakeOutput struct {
	Address                      string `json:"address" msgpack:"address"`
	DelegatedStakeSpendPublicKey []byte `json:"delegatedStakeSpendPublicKey" msgpack:"delegatedStakeSpendPublicKey"`
}

type GenesisDataAirDropType struct {
	Address        string `json:"address" msgpack:"address"`
	Amount         uint64 `json:"amount" msgpack:"amount"`
	SpendPublicKey []byte `json:"spendPublicKey" msgpack:"spendPublicKey"`
}

type GenesisDataType struct {
	Hash       []byte                    `json:"hash" msgpack:"hash"`             //32 byte
	KernelHash []byte                    `json:"kernelHash" msgpack:"kernelHash"` //32 byte
	Timestamp  uint64                    `json:"timestamp" msgpack:"timestamp"`
	Target     []byte                    `json:"target" msgpack:"target"` //32 byte
	AirDrops   []*GenesisDataAirDropType `json:"airDrops" msgpack:"airDrops"`
}

var genesisMainet = GenesisDataType{
	Hash:       helpers.DecodeHex(""),
	KernelHash: helpers.DecodeHex(""),
	Timestamp:  uint64(0),
	Target:     helpers.DecodeHex(""),
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
	KernelHash: helpers.DecodeHex("0000000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
	Timestamp:  uint64(time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC).Unix()),
	Target:     helpers.DecodeHex("0000000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
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
		MerkleHash:     cryptography.SHA3([]byte{}),
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

	var data []byte

	GenesisData.Hash = helpers.RandomBytes(cryptography.HashSize)
	GenesisData.Timestamp = uint64(time.Now().Unix()) //the reason is to forge first block fast in tests

	amount := 100 * config_stake.GetRequiredStake(0)
	for i := 1; i < len(v); i++ {

		if data, err = ioutil.ReadFile(v[i]); err != nil {
			return
		}

		delegatedStakeOutput := &DelegatedStakeOutput{}
		if err = json.Unmarshal(data, delegatedStakeOutput); err != nil {
			return
		}

		GenesisData.AirDrops = append(GenesisData.AirDrops, &GenesisDataAirDropType{
			Address:        delegatedStakeOutput.Address,
			Amount:         amount,
			SpendPublicKey: []byte{},
		})

	}

	//let's create 1000 zero wallets
	for i := 0; i < 1000; i++ {
		priv := addresses.GenerateNewPrivateKey()

		var addr *addresses.Address
		if addr, err = priv.GenerateAddress(true, nil, 0, nil); err != nil {
			return
		}

		GenesisData.AirDrops = append(GenesisData.AirDrops, &GenesisDataAirDropType{
			Address:        addr.EncodeAddr(),
			Amount:         0,
			SpendPublicKey: []byte{},
		})
	}

	if data, err = msgpack.Marshal(GenesisData); err != nil {
		return
	}

	if _, err = file.Write(data); err != nil {
		return
	}

	return
}

func createSimpleGenesis(walletGetFirstAddressForDevnetGenesisAirdrop func() (string, []byte, error)) (err error) {

	var file *os.File

	if globals.Arguments["--new-devnet"] == false {
		return errors.New("Genesis Data was not found and --new-devnet is missing")
	}

	GenesisData.Hash = helpers.RandomBytes(cryptography.HashSize)
	GenesisData.Timestamp = uint64(time.Now().Unix()) //the reason is to forge first block fast in tests

	address, delegatedStakePublicKey, err := walletGetFirstAddressForDevnetGenesisAirdrop()
	if err != nil {
		return
	}

	amount := 100 * config_stake.GetRequiredStake(0)
	GenesisData.AirDrops = append(GenesisData.AirDrops, &GenesisDataAirDropType{
		Address:        address,
		Amount:         amount,
		SpendPublicKey: delegatedStakePublicKey,
	})

	if file, err = os.OpenFile("./genesis.data", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return
	}
	defer file.Close()

	var data []byte
	if data, err = msgpack.Marshal(GenesisData); err != nil {
		return
	}

	if _, err = file.Write(data); err != nil {
		return
	}

	return
}

func GenesisInit(walletGetFirstAddressForDevnetGenesisAirdrop func() (string, []byte, error)) (err error) {

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
				if err = createSimpleGenesis(walletGetFirstAddressForDevnetGenesisAirdrop); err != nil {
					return
				}
			}

			if data, err = ioutil.ReadFile("./genesis.data"); err != nil {
				return
			}

		}

		if err = msgpack.Unmarshal(data, &GenesisData); err != nil {
			return
		}

	}

	if Genesis, err = CreateNewGenesisBlock(); err != nil {
		return
	}

	return
}
