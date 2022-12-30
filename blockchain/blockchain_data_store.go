package blockchain

import (
	"encoding/binary"
	"errors"
	"math"
	"math/big"
	"pandora-pay/helpers/advanced_buffers"
	"pandora-pay/helpers/msgpack"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

// chain must be locked before
func (chainData *BlockchainData) saveTotalDifficultyExtra(writer store_db_interface.StoreDBTransactionInterface) {
	key := "totalDifficulty" + strconv.FormatUint(chainData.Height, 10)

	bufferWriter := advanced_buffers.NewBufferWriter()
	bufferWriter.WriteUvarint(chainData.Timestamp)

	bytes := chainData.BigTotalDifficulty.Bytes()
	bufferWriter.WriteVariableBytes(bytes)

	writer.Put(key, bufferWriter.Bytes())
}

func (chainData *BlockchainData) LoadTotalDifficultyExtra(reader store_db_interface.StoreDBTransactionInterface, height uint64) (*big.Int, uint64, error) {
	if height < 0 {
		return nil, 0, errors.New("height is invalid")
	}

	buf := reader.Get("totalDifficulty" + strconv.FormatUint(height, 10))
	if buf == nil {
		return nil, 0, errors.New("Couldn't read difficulty from DB")
	}

	bufferReader := advanced_buffers.NewBufferReader(buf)
	timestamp, err := bufferReader.ReadUvarint()

	if err != nil {
		return nil, 0, err
	}

	bytes, err := bufferReader.ReadVariableBytes(math.MaxUint64)
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
	return msgpack.Unmarshal(chainInfoData, chainData)
}

func (chainData *BlockchainData) saveBlockchainHeight(writer store_db_interface.StoreDBTransactionInterface) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, chainData.Height)
	writer.Put("chainHeight", buf[:n])
	writer.Put("chainHash", chainData.Hash)
	writer.Put("chainPrevHash", chainData.PrevHash)
	writer.Put("chainKernelHash", chainData.KernelHash)
	writer.Put("chainPrevKernelHash", chainData.PrevKernelHash)
}

func (chainData *BlockchainData) saveBlockchainInfo(writer store_db_interface.StoreDBTransactionInterface) (err error) {
	var data []byte
	if data, err = msgpack.Marshal(chainData); err != nil {
		return
	}

	writer.Put("blockchainInfo_"+strconv.FormatUint(chainData.Height, 10), data)
	return nil
}

func (chainData *BlockchainData) saveBlockchain(writer store_db_interface.StoreDBTransactionInterface) error {
	marshal, err := msgpack.Marshal(chainData)
	if err != nil {
		return err
	}

	writer.Put("blockchainInfo", marshal)
	return nil
}
