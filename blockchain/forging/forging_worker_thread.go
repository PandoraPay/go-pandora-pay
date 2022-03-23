package forging

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"strconv"
	"sync/atomic"
	"time"
)

type ForgingSolution struct {
	timestamp     uint64
	address       *ForgingWalletAddress
	work          *forging_block_work.ForgingWork
	stakingAmount uint64
	stakingNonce  []byte
}

type ForgingWorkerThread struct {
	addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor
	hashes                  uint32
	index                   int
	workCn                  chan *forging_block_work.ForgingWork
	workerSolutionCn        chan *ForgingSolution
	addWalletAddressCn      chan *ForgingWalletAddress
	removeWalletAddressCn   chan string //publicKey
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

			if !bytes.Equal(threadAddr.stakingNoncePrevChainKernelHash, work.BlkComplete.PrevKernelHash) {
				uinput := append([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT), work.BlkComplete.PrevKernelHash[:]...)
				uinput = append(uinput, config_coins.NATIVE_ASSET_FULL...)
				uinput = append(uinput, strconv.Itoa(0)...)
				u := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(uinput)), threadAddr.walletAdr.privateKeyPoint)
				threadAddr.stakingNonce = cryptography.SHA3(u.EncodeCompressed())
				threadAddr.stakingNoncePrevChainKernelHash = work.BlkComplete.PrevKernelHash
			}

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

		timestamp = work.BlkTimestmap + 1

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
				if worker.computeStakingAmount(walletAddr, work) {
					if !walletsStakedUsed[walletAddr.walletAdr.publicKeyStr] {
						walletsStaked[walletAddr.walletAdr.publicKeyStr] = walletAddr
						walletsStakedTimestamp[walletAddr.walletAdr.publicKeyStr] = timestamp
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
					case publicKeyStr := <-worker.removeWalletAddressCn:
						removeWalletAddr(publicKeyStr)
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
							address.walletAdr,
							work,
							generics.Max(generics.Min(requireStakingAmount.Uint64()+1, address.stakingAmount), work.MinimumStake),
							address.stakingNonce,
						}

						worker.workerSolutionCn <- solution

						delete(walletsStaked, key)
						walletsStakedUsed[key] = true

					} /* else { // for debugging only
						gui.GUI.Log(base64.StdEncoding.EncodeToString(kernelHash), strconv.FormatUint(timestamp, 10 ))
					}*/

					walletsStakedTimestamp[key] += 1
					hashes++
				}

			}

			return false
		}()
		atomic.AddUint32(&worker.hashes, uint32(hashes))

		if hashes == 0 && !hasNewWork {
			time.Sleep(time.Duration(((timeLimitMs/1000+1)*1000 - timeLimitMs) * 1000000))
		}

	}

}

func createForgingWorkerThread(index int, workerSolutionCn chan *ForgingSolution, addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor) *ForgingWorkerThread {
	return &ForgingWorkerThread{
		addressBalanceDecryptor: addressBalanceDecryptor,
		index:                   index,
		workCn:                  make(chan *forging_block_work.ForgingWork),
		workerSolutionCn:        workerSolutionCn,
		addWalletAddressCn:      make(chan *ForgingWalletAddress),
		removeWalletAddressCn:   make(chan string),
	}
}
