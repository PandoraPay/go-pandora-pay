package forging

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"sync/atomic"
	"time"
)

type ForgingSolution struct {
	timestamp               uint64
	publicKey               []byte
	decryptedStakingBalance uint64
	blkComplete             *block_complete.BlockComplete
	stakingAmount           uint64
	stakingNonce            []byte
}

type ForgingWorkerThread struct {
	hashes                uint32
	index                 int
	workCn                chan *forging_block_work.ForgingWork
	workerSolutionCn      chan *ForgingSolution
	addWalletAddressCn    chan *ForgingWalletAddress
	removeWalletAddressCn chan string //publicKey
}

type ForgingWorkerThreadAddress struct {
	walletAdr                       *ForgingWalletAddress
	stakingAmount                   uint64
	stakingNonce                    []byte
	stakingNoncePrevChainKernelHash []byte
}

func (worker *ForgingWorkerThread) computeStakingAmount(threadAddr *ForgingWorkerThreadAddress, work *forging_block_work.ForgingWork) bool {

	if threadAddr.walletAdr.account != nil && threadAddr.walletAdr.privateKey != nil {

		if threadAddr.walletAdr.decryptedStakingBalance >= work.MinimumStake {
			threadAddr.stakingAmount = threadAddr.walletAdr.decryptedStakingBalance
			return true
		}

	}

	threadAddr.stakingAmount = 0
	return false
}

/**
"Staking multiple wallets simultaneously"
*/
func (worker *ForgingWorkerThread) forge() {

	var work *forging_block_work.ForgingWork

	var timestamp, localTimestamp uint64
	var serialized []byte
	var ok bool
	var n int
	var hashes int32
	buf := make([]byte, binary.MaxVarintLen64)

	wallets := make(map[string]*ForgingWorkerThreadAddress)
	walletsStaked := make(map[string]*ForgingWorkerThreadAddress)
	walletsStakedTimestamp := make(map[string]uint64)
	walletsStakedUsed := make(map[string]bool)

	waitCn := make(chan struct{})
	waitCnClosed := false

	validateWork := func() {
		if work == nil || len(walletsStaked) == 0 {
			if waitCnClosed {
				waitCn = make(chan struct{})
				waitCnClosed = false
			}
		} else {
			if !waitCnClosed {
				close(waitCn)
				waitCnClosed = true
			}
		}
	}

	newWork := func(newWorkReceived *forging_block_work.ForgingWork) {

		work = newWorkReceived

		serialized = helpers.CloneBytes(work.BlkSerialized)

		timestamp = work.BlkTimestmap

		n = binary.PutUvarint(buf, timestamp)

		walletsStaked = make(map[string]*ForgingWorkerThreadAddress)
		walletsStakedTimestamp = make(map[string]uint64)
		walletsStakedUsed = make(map[string]bool)

		for _, walletAddr := range wallets {
			if worker.computeStakingAmount(walletAddr, work) {
				walletsStaked[walletAddr.walletAdr.publicKeyStr] = walletAddr
				walletsStakedTimestamp[walletAddr.walletAdr.publicKeyStr] = timestamp
			}
		}

		validateWork()
	}

	newWalletAddress := func(newWalletAddr *ForgingWalletAddress) {
		walletAddr := wallets[newWalletAddr.publicKeyStr]
		if walletAddr == nil {
			walletAddr = &ForgingWorkerThreadAddress{ //making sure the has a copy
				newWalletAddr, //already it is cloned
				0,
				nil,
				nil,
			}
			wallets[newWalletAddr.publicKeyStr] = walletAddr
		} else {
			walletAddr.walletAdr = newWalletAddr
		}

		if work != nil {
			if walletAddr.walletAdr.chainHash == nil || bytes.Equal(walletAddr.walletAdr.chainHash, work.BlkComplete.PrevHash) {
				oldDecryptedStakingBalance := walletAddr.stakingAmount
				if worker.computeStakingAmount(walletAddr, work) {
					if !walletsStakedUsed[walletAddr.walletAdr.publicKeyStr] || walletAddr.stakingAmount != oldDecryptedStakingBalance {
						walletsStaked[walletAddr.walletAdr.publicKeyStr] = walletAddr
						walletsStakedTimestamp[walletAddr.walletAdr.publicKeyStr] = timestamp
						delete(walletsStakedUsed, walletAddr.walletAdr.publicKeyStr)
					}
				} else {
					delete(walletsStaked, walletAddr.walletAdr.publicKeyStr)
				}
			}
		}

		validateWork()
	}

	removeWalletAddr := func(publicKeyStr string) {
		if wallets[publicKeyStr] != nil {
			delete(wallets, publicKeyStr)
			delete(walletsStaked, publicKeyStr)
			delete(walletsStakedTimestamp, publicKeyStr)
		}
		validateWork()
	}

	for {

		select {
		case newWorkReceived := <-worker.workCn: //or the work was changed meanwhile
			newWork(newWorkReceived)
			continue
		case newWalletAddr := <-worker.addWalletAddressCn:
			newWalletAddress(newWalletAddr)
			continue
		case publicKeyStr := <-worker.removeWalletAddressCn:
			removeWalletAddr(publicKeyStr)
			continue
		case <-waitCn:
		}

		if len(walletsStaked) == 0 || work == nil {
			validateWork()
			continue
		}

		timeLimitMs := time.Now().UnixNano()/1000000 + config.NETWORK_TIMESTAMP_DRIFT_MAX_INT*1000
		timeLimit := uint64(timeLimitMs / 1000)

		hashes = 0

		hasNewWork := func() bool {

			for key, address := range walletsStaked {
				localTimestamp, ok = walletsStakedTimestamp[key]
				if ok && localTimestamp < timeLimit {

					select {
					case newWorkReceived := <-worker.workCn: //or the work was changed meanwhile
						newWork(newWorkReceived)
						return true
					case newWalletAddr := <-worker.addWalletAddressCn:
						newWalletAddress(newWalletAddr)
						if key == newWalletAddr.publicKeyStr {
							goto done
						}
					case publicKeyStr := <-worker.removeWalletAddressCn:
						removeWalletAddr(publicKeyStr)
						if key == publicKeyStr {
							goto done
						}
					default:
					}

					n2 := binary.PutUvarint(buf, localTimestamp)

					if n2 != n {
						newSerialized := make([]byte, len(serialized)-n+n2)
						copy(newSerialized, serialized[:n-32])
						serialized = newSerialized
						n = n2
					}

					//optimized POS
					copy(serialized[len(serialized)-32-n2:len(serialized)-32], buf)
					copy(serialized[len(serialized)-32:], address.stakingNonce)

					kernelHash := cryptography.SHA3(serialized)

					kernel := new(big.Int).Div(new(big.Int).SetBytes(kernelHash), new(big.Int).SetUint64(address.stakingAmount))

					if kernel.Cmp(work.Target) <= 0 {

						requireStakingAmount := new(big.Int).Div(new(big.Int).SetBytes(kernelHash), work.Target)

						gui.GUI.Log("forged", worker.index, " -> ", work.BlkHeight, work.BlkComplete.PrevHash, address.walletAdr.decryptedStakingBalance)

						solution := &ForgingSolution{
							localTimestamp,
							address.walletAdr.publicKey,
							address.walletAdr.decryptedStakingBalance,
							work.BlkComplete,
							generics.Max(generics.Min(requireStakingAmount.Uint64()+1, address.stakingAmount), work.MinimumStake),
							address.stakingNonce,
						}

						select {
						case newWorkReceived := <-worker.workCn: //or the work was changed meanwhile
							newWork(newWorkReceived)
							return true
						case newWalletAddr := <-worker.addWalletAddressCn:
							newWalletAddress(newWalletAddr)
							if key == newWalletAddr.publicKeyStr { // in case it was deleted
								goto done
							}
						case publicKeyStr := <-worker.removeWalletAddressCn:
							removeWalletAddr(publicKeyStr)
							if key == publicKeyStr { // in case it was deleted
								goto done
							}
						case worker.workerSolutionCn <- solution:
							delete(walletsStaked, key)
							walletsStakedUsed[key] = true
						}

					} /* else { // for debugging only
						gui.GUI.Log(base64.StdEncoding.EncodeToString(kernelHash), strconv.FormatUint(timestamp, 10 ))
					}*/

					walletsStakedTimestamp[key] += 1
					hashes++
				}

			done:
			}

			return false
		}()
		atomic.AddUint32(&worker.hashes, uint32(hashes))

		if hashes == 0 && !hasNewWork {
			time.Sleep(time.Duration(((timeLimitMs/1000+1)*1000 - timeLimitMs) * 1000000))
		}

	}

}

func createForgingWorkerThread(index int, workerSolutionCn chan *ForgingSolution) *ForgingWorkerThread {
	return &ForgingWorkerThread{
		index:                 index,
		workCn:                make(chan *forging_block_work.ForgingWork),
		workerSolutionCn:      workerSolutionCn,
		addWalletAddressCn:    make(chan *ForgingWalletAddress),
		removeWalletAddressCn: make(chan string),
	}
}
