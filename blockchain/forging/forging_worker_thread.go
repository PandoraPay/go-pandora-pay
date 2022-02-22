package forging

import (
	"context"
	"encoding/binary"
	"math/big"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
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
	hashes                uint32
	index                 int
	workCn                chan *forging_block_work.ForgingWork
	workerSolutionCn      chan *ForgingSolution
	addWalletAddressCn    chan *ForgingWalletAddress
	removeWalletAddressCn chan string //publicKey
}

type ForgingWorkerThreadAddress struct {
	walletAdr     *ForgingWalletAddress
	stakingAmount uint64
	stakingNonce  []byte
}

func (threadAddr *ForgingWorkerThreadAddress) computeStakingAmount(height uint64, prevChainKernelHash []byte) bool {

	if threadAddr.walletAdr.account != nil && threadAddr.walletAdr.privateKey != nil {

		if threadAddr.walletAdr.account != nil {

			stakingAmountBalance := threadAddr.walletAdr.account.DelegatedStake.ComputeDelegatedStakeAvailable(threadAddr.walletAdr.account.Balance.Amount, height)
			if stakingAmountBalance != nil {
				threadAddr.stakingAmount, _ = threadAddr.walletAdr.privateKey.DecryptBalance(stakingAmountBalance, threadAddr.stakingAmount, context.Background(), func(string) {})
			}
		}

		if threadAddr.stakingAmount >= config_stake.GetRequiredStake(height) {

			uinput := append([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT), prevChainKernelHash[:]...)
			uinput = append(uinput, config_coins.NATIVE_ASSET_FULL...)
			uinput = append(uinput, strconv.Itoa(0)...)
			u := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(uinput)), threadAddr.walletAdr.privateKeyPoint) // this should be moved to generate proof
			threadAddr.stakingNonce = cryptography.SHA3(u.EncodeCompressed())

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

	var timestampMs int64
	var timestamp, blkHeight uint64
	var serialized []byte
	var n int
	buf := make([]byte, binary.MaxVarintLen64)

	wallets := make(map[string]*ForgingWorkerThreadAddress)
	walletsStakable := make(map[string]*ForgingWorkerThreadAddress)

	waitCn := make(chan struct{})
	waitCnClosed := false

	validateWork := func() {
		if work == nil || len(walletsStakable) == 0 {
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

		blkHeight = work.BlkHeight
		timestamp = work.BlkTimestmap + 1
		timestampMs = int64(timestamp) * 1000

		n = binary.PutUvarint(buf, timestamp)

		walletsStakable = make(map[string]*ForgingWorkerThreadAddress)
		for _, walletAddr := range wallets {
			if walletAddr.computeStakingAmount(blkHeight, work.BlkComplete.PrevKernelHash) {
				walletsStakable[walletAddr.walletAdr.publicKeyStr] = walletAddr
			}
		}

		validateWork()
	}

	for {

		select {
		case newWorkReceived := <-worker.workCn: //or the work was changed meanwhile
			newWork(newWorkReceived)
			continue
		case newWalletAddr := <-worker.addWalletAddressCn:
			walletAddr := wallets[newWalletAddr.publicKeyStr]
			if walletAddr == nil {
				walletAddr = &ForgingWorkerThreadAddress{ //making sure the has a copy
					newWalletAddr, //already it is copied
					0,
					nil,
				}
				wallets[newWalletAddr.publicKeyStr] = walletAddr
			} else {
				walletAddr.walletAdr = newWalletAddr
			}

			if work != nil {
				if walletAddr.computeStakingAmount(blkHeight, work.BlkComplete.PrevKernelHash) {
					walletsStakable[walletAddr.walletAdr.publicKeyStr] = walletAddr
				} else {
					delete(walletsStakable, walletAddr.walletAdr.publicKeyStr)
				}
			}

			validateWork()
			continue
		case publicKeyStr := <-worker.removeWalletAddressCn:
			if wallets[publicKeyStr] != nil {
				delete(wallets, publicKeyStr)
				delete(walletsStakable, publicKeyStr)
			}
			validateWork()
			continue
		case <-waitCn:
		}

		timeLimitMs := time.Now().UnixNano()/1000000 + config.NETWORK_TIMESTAMP_DRIFT_MAX_INT*1000

		if timestampMs > timeLimitMs {
			time.Sleep(time.Millisecond * time.Duration(timestampMs-timeLimitMs))
			continue
		}

		timeLimit := uint64(timeLimitMs / 1000)
		//forge with my wallets
		diff := 0
		if timeLimit > timestamp {
			diff = int(generics.Min(timeLimit-timestamp, 20))
		}

		func() {
			for i := 0; i <= diff; i++ {
				for key, address := range walletsStakable {

					select {
					case newWorkReceived := <-worker.workCn: //or the work was changed meanwhile
						newWork(newWorkReceived)
						return
					default:
					}

					n2 := binary.PutUvarint(buf, timestamp)

					if n2 != n {
						newSerialized := make([]byte, len(serialized)-n+n2)
						copy(newSerialized, serialized[:-n-32])
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

						solution := &ForgingSolution{
							timestamp,
							address.walletAdr,
							work,
							generics.Min(requireStakingAmount.Uint64()+1, address.stakingAmount),
							address.stakingNonce,
						}

						select {
						case worker.workerSolutionCn <- solution:
						case newWorkReceived := <-worker.workCn: //or the work was changed meanwhile
							newWork(newWorkReceived)
							return
						}

						delete(walletsStakable, key)

					} /* else { // for debugging only
						gui.GUI.Log(base64.StdEncoding.EncodeToString(kernelHash), strconv.FormatUint(timestamp, 10 ))
					}*/

				}

				timestamp += 1
				timestampMs += 1000

			}
			atomic.AddUint32(&worker.hashes, uint32((diff+1)*len(walletsStakable)))
		}()

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
