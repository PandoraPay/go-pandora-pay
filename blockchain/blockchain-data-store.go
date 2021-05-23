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
	key := []byte("totalDifficulty" + strconv.FormatUint(chainData.Height, 10))

	bufferWriter := helpers.NewBufferWriter()
	bufferWriter.WriteUvarint(chainData.Timestamp)

	bytes := chainData.BigTotalDifficulty.Bytes()
	bufferWriter.WriteUvarint(uint64(len(bytes)))
	bufferWriter.Write(bytes)

	return writer.Put(key, bufferWriter.Bytes())
}

func (chainData *BlockchainData) LoadTotalDifficultyExtra(reader store_db_interface.StoreDBTransactionInterface, height uint64) (difficulty *big.Int, timestamp uint64, err error) {
	if height < 0 {
		err = errors.New("height is invalid")
		return
	}
	key := []byte("totalDifficulty" + strconv.FormatUint(height, 10))

	buf := reader.Get(key)
	if buf == nil {
		err = errors.New("Couldn't read difficulty from DB")
		return
	}

	bufferReader := helpers.NewBufferReader(buf)
	if timestamp, err = bufferReader.ReadUvarint(); err != nil {
		return
	}

	length, err := bufferReader.ReadUvarint()
	if err != nil {
		return
	}

	bytes, err := bufferReader.ReadBytes(int(length))
	if err != nil {
		return
	}

	difficulty = new(big.Int).SetBytes(bytes)

	return
}

func (chainData *BlockchainData) loadBlockchainInfo(reader store_db_interface.StoreDBTransactionInterface, height uint64) error {
	chainInfoData := reader.Get([]byte("blockchainInfo_" + strconv.FormatUint(height, 10)))
	if chainInfoData == nil {
		return errors.New("Chain not found")
	}
	return json.Unmarshal(chainInfoData, chainData)
}

func (chainData *BlockchainData) saveBlockchainHeight(writer store_db_interface.StoreDBTransactionInterface) (err error) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, chainData.Height)
	return writer.Put([]byte("chainHeight"), buf[:n])
}

func (chainData *BlockchainData) saveBlockchainInfo(writer store_db_interface.StoreDBTransactionInterface) (err error) {
	var data []byte
	if data, err = json.Marshal(chainData); err != nil {
		return
	}

	return writer.Put([]byte("blockchainInfo_"+strconv.FormatUint(chainData.Height, 10)), data)
}

func (chainData *BlockchainData) saveBlockchain(writer store_db_interface.StoreDBTransactionInterface) error {
	marshal, err := json.Marshal(chainData)
	if err != nil {
		return err
	}

	return writer.Put([]byte("blockchainInfo"), marshal)
}
