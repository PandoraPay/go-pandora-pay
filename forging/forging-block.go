package forging

import (
	"pandora-pay/block"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/crypto"
	"pandora-pay/wallet"
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

		return &block, nil
	}

}

//inside a thread
func forge(block *block.Block, threadIndex int) {

	serialized := block.SerializeBlock(false, false, false, false)

	addresses := wallet.GetAddresses()
	//forge with my wallets
	for i := 0; i < len(addresses); i++ {
		if i%threadIndex == 0 {
			finalSerialized := append(serialized, addresses[i].Address.PublicKey...)
			kernelHash := crypto.SHA3Hash(finalSerialized)

		}
	}

}
