package forging

import (
	"pandora-pay/block"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/crypto"
	"time"
)

func createNextBlock(height uint64) (*block.Block, error) {

	if height == 0 {
		return genesis.CreateGenesisBlock()
	} else {
		now := time.Now()
		var blockHeader = block.BlockHeader{
			MajorVersion: 0,
			MinorVersion: 0,
			Timestamp:    uint64(now.Unix()),
			Height:       blockchain.Chain.Height,
		}
		var block = block.Block{
			BlockHeader:    blockHeader,
			MerkleHash:     crypto.SHA3Hash([]byte{}),
			PrevHash:       blockchain.Chain.Hash,
			PrevKernelHash: blockchain.Chain.KernelHash,
		}
	}

}
