package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
	"sync"
	"time"
)

type Blockchain struct {
	Hash       helpers.Hash
	KernelHash helpers.Hash
	Height     uint64
	Timestamp  uint64

	Difficulty         uint64
	Target             *big.Int
	BigTotalDifficulty *big.Int

	Sync bool `json:"-"`

	UpdateChannel chan int `json:"-"`

	mutex        sync.Mutex `json:"-"`
	sync.RWMutex `json:"-"`
}

var Chain Blockchain

func (chain *Blockchain) AddBlocks(blocksComplete []*block.BlockComplete) (result bool, err error) {

	result = false
	if len(blocksComplete) == 0 {
		err = errors.New("Blocks length is ZERO")
		return
	}

	//avoid processing the same function twice
	chain.mutex.Lock()
	defer chain.mutex.Unlock()

	var wasChainLocked bool

	gui.Log(fmt.Sprintf("Including blocks %d ... %d", chain.Height, chain.Height+uint64(len(blocksComplete))))

	var newChain = Blockchain{
		Hash:               chain.Hash,
		KernelHash:         chain.KernelHash,
		Height:             chain.Height,
		Timestamp:          chain.Timestamp,
		Difficulty:         chain.Difficulty,
		Target:             chain.Target,
		BigTotalDifficulty: chain.BigTotalDifficulty,
	}

	var accs *accounts.Accounts

	err = store.StoreBlockchain.DB.Update(func(tx *bolt.Tx) (err error) {

		if accs, err = accounts.CreateNewAccounts(tx); err != nil {
			return
		}

		writer := tx.Bucket([]byte("Chain"))

		var prevBlk = &block.Block{}
		if blocksComplete[0].Block.Height == 0 {
			prevBlk = genesis.Genesis
		} else {
			if prevBlk, err = loadBlock(writer, newChain.Hash); err != nil {
				return
			}
		}

		//let's filter existing blocks
		for i := len(blocksComplete) - 1; i >= 0; i-- {

			blkComplete := blocksComplete[i]

			if blkComplete.Block.Height > chain.Height {
				var hash helpers.Hash
				hash, err = newChain.loadBlockHash(writer, blkComplete.Block.Height)
				if err != nil {
					return err
				}

				hash2 := blkComplete.Block.ComputeHash()
				if bytes.Equal(hash[:], hash2[:]) {
					blocksComplete = append(blocksComplete[:i], blocksComplete[i+1:]...)
				}
			}
		}

		if blocksComplete[0].Block.Height != newChain.Height {
			return errors.New("First Block has is not matching")
		}

		if !bytes.Equal(blocksComplete[0].Block.PrevHash[:], newChain.Hash[:]) {
			return errors.New("First block hash is not matching chain hash")
		}

		if !bytes.Equal(blocksComplete[0].Block.PrevKernelHash[:], newChain.KernelHash[:]) {
			return errors.New("First block kernel hash is not matching chain prev kerneh lash")
		}

		for i, blkComplete := range blocksComplete {

			hash := blkComplete.Block.ComputeHash()
			kernelHash := blkComplete.Block.ComputeKernelHash()

			if difficulty.CheckKernelHashBig(kernelHash, Chain.Target) != true {
				return errors.New("KernelHash Difficulty is not met")
			}

			//already verified for i == 0
			if i > 0 {

				prevHash := prevBlk.ComputeHash()
				if !bytes.Equal(blkComplete.Block.PrevHash[:], prevHash[:]) {
					return errors.New("PrevHash doesn't match Genesis prevHash")
				}

				prevKernelHash := prevBlk.ComputeKernelHash()
				if !bytes.Equal(blkComplete.Block.PrevKernelHash[:], prevKernelHash[:]) {
					return errors.New("PrevHash doesn't match Genesis prevKernelHash")
				}

			}

			if blkComplete.Block.VerifySignature() != true {
				return errors.New("Forger Signature is invalid!")
			}

			if blkComplete.Block.BlockHeader.Version != 0 {
				return errors.New("Invalid Version Version")
			}

			if blkComplete.Block.Timestamp < newChain.Timestamp {
				return errors.New("Timestamp has to be greather than the last timestmap")
			}

			if blkComplete.Block.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
				return errors.New("Timestamp is too much into the future")
			}

			if blkComplete.VerifyMerkleHash() != true {
				return errors.New("Verify Merkle Hash failed")
			}

			if err = blkComplete.Block.IncludeBlock(accs); err != nil {
				return
			}

			if err = saveBlock(writer, blkComplete, hash); err != nil {
				return
			}

			if err = accs.Commit(); err != nil {
				return
			}

			newChain.Hash = hash
			newChain.KernelHash = kernelHash
			newChain.Timestamp = blkComplete.Block.Timestamp

			bigKernelHash := difficulty.HashToBig(kernelHash)
			difficultyKernelHash := difficulty.ConvertDifficultyBigToUInt64(bigKernelHash)

			if newChain.Target, err = newChain.computeNextDifficultyBig(writer); err != nil {
				return
			}

			newChain.BigTotalDifficulty = new(big.Int).Add(newChain.BigTotalDifficulty, new(big.Int).SetUint64(difficultyKernelHash))
			if err = newChain.saveTotalDifficultyExtra(writer); err != nil {
				return
			}

			newChain.Height += 1

		}

		err = saveBlockchain(writer, &newChain)

		chain.Lock()
		wasChainLocked = true
		chain.Height = newChain.Height
		chain.Hash = newChain.Hash
		chain.KernelHash = newChain.KernelHash
		chain.Timestamp = newChain.Timestamp
		chain.Target = newChain.Target
		chain.BigTotalDifficulty = newChain.BigTotalDifficulty

		return
	})

	if wasChainLocked {
		chain.Unlock()
	}
	if err != nil {
		return
	}

	gui.Log(fmt.Sprintf("Including blocks SUCCESS %s", hex.EncodeToString(chain.Hash[:])))
	updateChainInfo()

	chain.UpdateChannel <- 1 //sending 1

	go chain.createBlockForForging()

	result = true
	return

}

func (chain *Blockchain) createBlockForForging() {

	var err error

	var nextBlock *block.BlockComplete
	if nextBlock, err = Chain.createNextBlockComplete(); err == nil {
		gui.Error("Error creating next block", err)
	}

	forging.Forging.RestartForgingWorkers(nextBlock, chain.Target)
}

func BlockchainInit() {

	gui.Info("Blockchain init...")

	genesis.GenesisInit()

	success, err := loadBlockchain()
	if err != nil {
		gui.Fatal("Loading a blockchain info raised an error", err)
	}

	if !success {
		Chain.Height = 0
		Chain.Hash = genesis.GenesisData.Hash
		Chain.KernelHash = genesis.GenesisData.KernelHash
		Chain.Difficulty = genesis.GenesisData.Difficulty
		Chain.Target = difficulty.ConvertDifficultyToBig(Chain.Difficulty)
		Chain.BigTotalDifficulty = new(big.Int).SetUint64(0)
	}
	Chain.UpdateChannel = make(chan int)
	updateChainInfo()

	Chain.Sync = false

	forging.ForgingInit()

	go func() {

		for {
			_ = <-forging.Forging.SolutionChannel

			var array []*block.BlockComplete
			array = append(array, forging.Forging.BlkComplete)

			result, err := Chain.AddBlocks(array)
			if err == nil && result {
				gui.Info("Block was forged! " + strconv.FormatUint(forging.Forging.BlkComplete.Block.Height, 10))
			} else {
				gui.Error("Error forging block "+strconv.FormatUint(forging.Forging.BlkComplete.Block.Height, 10), err)
			}

		}

	}()

	go Chain.createBlockForForging()

}

func updateChainInfo() {
	gui.InfoUpdate("Blocks", strconv.FormatUint(Chain.Height, 10))
	gui.InfoUpdate("Chain Hash", hex.EncodeToString(Chain.Hash[:]))
	gui.InfoUpdate("Chain Diff", Chain.Target.String())
}
