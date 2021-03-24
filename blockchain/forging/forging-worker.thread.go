package forging

import (
	"encoding/binary"
	"math/big"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"sync"
	"time"
)

type ForgingWork struct {
	blkComplete *block_complete.BlockComplete
	target      *big.Int
}

type ForgingSolution struct {
	timestamp uint64
	address   *ForgingWalletAddress
	work      *ForgingWork
}
type ForgingWalletAddressRequired struct {
	publicKeyHash []byte //20 byte
	wallet        *ForgingWalletAddress
	stakingAmount uint64
}

/**
"Staking multiple wallets simulateneoly"
*/
func forge(
	wg *sync.WaitGroup,
	work *ForgingWork, // SAFE READ ONLY
	workChannel <-chan *ForgingWork, // SAFE
	solutionChannel chan *ForgingSolution, // SAFE
	wallets []*ForgingWalletAddressRequired, // SAFE READ ONLY
) {

	buf := make([]byte, binary.MaxVarintLen64)

	defer wg.Done()

	height := work.blkComplete.Block.Height
	serialized := work.blkComplete.Block.SerializeForForging()
	n := binary.PutUvarint(buf, work.blkComplete.Block.Timestamp)

	serialized = serialized[:len(serialized)-n-20]
	timestamp := work.blkComplete.Block.Timestamp + 1

	for {

		timeNow := uint64(time.Now().Unix()) + config.NETWORK_TIMESTAMP_DRIFT_MAX

		select {
		case <-workChannel: //or the work was changed meanwhile
		case <-solutionChannel: //someone published a solution
			return
		default:
			if timestamp > timeNow {
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}

		//forge with my wallets
		diff := timeNow - timestamp
		if diff > 20 {
			diff = 20
		}
		for i := uint64(0); i <= diff; i++ {
			for _, address := range wallets {

				n = binary.PutUvarint(buf, timestamp)
				serialized = append(serialized, buf[:n]...)
				serialized = append(serialized, address.publicKeyHash...)
				kernelHash := cryptography.SHA3Hash(serialized)

				if height > 0 {
					kernelHash = cryptography.ComputeKernelHash(kernelHash, address.stakingAmount)
				}

				if difficulty.CheckKernelHashBig(kernelHash, work.target) {

					solutionChannel <- &ForgingSolution{
						timestamp: timestamp,
						address:   address.wallet,
						work:      work,
					}
					close(solutionChannel)
					return

				} else {
					//for debugging only
					//gui.Log(hex.EncodeToString(kernelHash))
				}

				serialized = serialized[:len(serialized)-n-20]
			}

			timestamp += 1
		}
	}

}
