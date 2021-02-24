package blockchain

import (
	"bytes"
	"errors"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/block"
	"pandora-pay/block/difficulty"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/store"
	"sync"
	"time"
)

type Blockchain struct {
	Hash          crypto.Hash
	KernelHash    crypto.Hash
	Height        uint64
	Difficulty    uint64
	BigDifficulty *big.Int //named also as target

	Sync bool

	sync.RWMutex
}

var Chain Blockchain

func (chain *Blockchain) AddBlocks(blocksComplete []*block.BlockComplete) (result bool, err error) {

	result = false
	if len(blocksComplete) == 0 {
		err = errors.New("Blocks length is ZERO")
		return
	}

	chain.Lock()
	defer chain.Unlock() //when the function exists

	err = store.StoreBlockchain.DB.Update(func(tx *bolt.Tx) (err error) {

		var writer *bolt.Bucket
		writer, err = tx.CreateBucketIfNotExists([]byte("Chain"))
		if err != nil {
			return
		}

		var prevBlk = &block.Block{}
		if blocksComplete[0].Block.Height == 0 {
			prevBlk = genesis.Genesis
		} else {
			prevBlk, err = LoadBlock(writer, chain.Hash)
			if err != nil {
				return
			}
		}

		if !bytes.Equal(blocksComplete[0].Block.PrevHash[:], chain.Hash[:]) {
			err = errors.New("First block hash is not matching chain hash")
			return
		}

		if !bytes.Equal(blocksComplete[0].Block.PrevKernelHash[:], chain.KernelHash[:]) {
			err = errors.New("First block kernel hash is not matching chain prev kerneh lash")
			return
		}

		for i, blkComplete := range blocksComplete {

			if difficulty.CheckKernelHashBig(blkComplete.Block.ComputeKernelHash(), Chain.BigDifficulty) != true {
				err = errors.New("KernelHash Difficulty is not met")
				return
			}

			//already verified for i == 0
			if i > 0 {

				prevHash := prevBlk.ComputeHash()
				if !bytes.Equal(blkComplete.Block.PrevHash[:], prevHash[:]) {
					err = errors.New("PrevHash doesn't match Genesis prevHash")
					return
				}

				prevKernelHash := prevBlk.ComputeKernelHash()
				if !bytes.Equal(blkComplete.Block.PrevKernelHash[:], prevKernelHash[:]) {
					err = errors.New("PrevHash doesn't match Genesis prevKernelHash")
					return
				}

			}

			if blkComplete.Block.VerifySignature() != true {
				err = errors.New("Forger Signature is invalid!")
				return
			}

			if blkComplete.Block.BlockHeader.Version != 0 {
				err = errors.New("Invalid Version Version")
				return
			}

			if blkComplete.Block.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
				err = errors.New("Timestamp is too much into the future")
				return
			}

		}

		return
	})

	if err != nil {
		return
	}

	result = true
	return

}

func BlockchainInit() {

	gui.Info("Blockchain init...")

	genesis.GenesisInit()

	Chain.Height = 0
	Chain.Hash = genesis.GenesisData.Hash
	Chain.KernelHash = genesis.GenesisData.KernelHash
	Chain.Difficulty = genesis.GenesisData.Difficulty
	Chain.BigDifficulty = difficulty.ConvertDifficultyToBig(Chain.Difficulty)
	Chain.Sync = false

}
