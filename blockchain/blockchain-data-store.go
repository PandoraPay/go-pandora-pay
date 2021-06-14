package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

//chain must be locked before
func (chainData *BlockchainData) saveTotalDifficultyExtra(writer store_db_interface.StoreDBTransactionInterface) error {
	key := "totalDifficulty" + strconv.FormatUint(chainData.Height, 10)

	bufferWriter := helpers.NewBufferWriter()
	bufferWriter.WriteUvarint(chainData.Timestamp)

	bytes := chainData.BigTotalDifficulty.Bytes()
	bufferWriter.WriteUvarint(uint64(len(bytes)))
	bufferWriter.Write(bytes)

	return writer.Put(key, bufferWriter.Bytes())
}

func (chainData *BlockchainData) LoadTotalDifficultyExtra(reader store_db_interface.StoreDBTransactionInterface, height uint64) (*big.Int, uint64, error) {
	if height < 0 {
		return nil, 0, errors.New("height is invalid")
	}

	buf := reader.Get("totalDifficulty" + strconv.FormatUint(height, 10))
	if buf == nil {
		return nil, 0, errors.New("Couldn't read difficulty from DB")
	}

	bufferReader := helpers.NewBufferReader(buf)
	timestamp, err := bufferReader.ReadUvarint()

	if err != nil {
		return nil, 0, err
	}

	length, err := bufferReader.ReadUvarint()
	if err != nil {
		return nil, 0, err
	}

	bytes, err := bufferReader.ReadBytes(int(length))
	if err != nil {
		return nil, 0, err
	}

	return new(big.Int).SetBytes(bytes), timestamp, nil
}

func (chainData *BlockchainData) loadBlockchainInfo(reader store_db_interface.StoreDBTransactionInterface, height uint64) error {
	chainInfoData := reader.Get("blockchainInfo_" + strconv.FormatUint(height, 10))
	if chainInfoData == nil {
		return errors.New("Chain not found")
	}
	return json.Unmarshal(chainInfoData, chainData)
}

func (chainData *BlockchainData) saveBlockchainHeight(writer store_db_interface.StoreDBTransactionInterface) (err error) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, chainData.Height)
	return writer.Put("chainHeight", buf[:n])
}

func (chainData *BlockchainData) saveBlockchainInfo(writer store_db_interface.StoreDBTransactionInterface) (err error) {
	var data []byte
	if data, err = json.Marshal(chainData); err != nil {
		return
	}

	return writer.Put("blockchainInfo_"+strconv.FormatUint(chainData.Height, 10), data)
}

func (chainData *BlockchainData) saveBlockchain(writer store_db_interface.StoreDBTransactionInterface) error {
	marshal, err := json.Marshal(chainData)
	if err != nil {
		return err
	}

	return writer.Put("blockchainInfo", marshal)
}
