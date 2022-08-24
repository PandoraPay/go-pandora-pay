package balance_decryptor

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"runtime"
	"sort"
	"strconv"
)

// this file implements balance decoder whih has to be bruteforced
// balance is a 64 bit field and total effort is 2^64
// but in reality, the balances are well distributed with the expectation that no one will ever be able to collect over 2 ^ 40
// However, 2^40 is solvable in less than 1 sec on a single core, with around 8 MB RAM, see below
// these tables are sharable between wallets

//type PreComputeTable [16*1024*1024]uint64 // each table is 128 MB in size

type PreComputeTable []uint64 // each table is 2^TABLE_SIZE * 8  bytes  in size
// 2^15 * 8 = 32768 *8 = 256 Kib
// 2^16 * 8 =            512 Kib
// 2^17 * 8 =              1 Mib
// 2^18 * 8 =              2 Mib
// 2^19 * 8 =              4 Mib
// 2^20 * 8 =              8 Mib
// 2^21 * 8 =              16 Mib
// 2^22 * 8 =              32 Mib

type LookupTable []PreComputeTable // default is 1 table, if someone who owns 2^48 coins or more needs more speed, it is possible

// IntSlice attaches the methods of Interface to []int, sorting in increasing order.

//type UintSlice []uint64

func (p PreComputeTable) Len() int           { return len(p) }
func (p PreComputeTable) Less(i, j int) bool { return p[i] < p[j] }
func (p PreComputeTable) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// with some more smartness table can be condensed more to contain 16.3% more entries within the same size
func createLookupTable(count, table_size int, tableComputedCn chan *LookupTable, readyCn chan struct{}, ctx context.Context, statusCallback func(string)) {

	t := make([]PreComputeTable, count, count)

	if table_size&0xff != 0 {
		panic("table size must be multiple of 256")
	}

	//terminal := isatty.IsTerminal(os.Stdout.Fd())

	var acc bn256.G1 // avoid allocations every loop
	acc.ScalarMult(crypto.G, new(big.Int).SetUint64(0))

	small_table := make([]*bn256.G1, 256, 256)
	for k := range small_table {
		small_table[k] = new(bn256.G1)
	}

	var compressed [33]byte

	for i := range t {
		t[i] = make([]uint64, table_size, table_size)

		for j := 0; j < table_size; j += 256 {

			for k := range small_table {
				small_table[k].Set(&acc)
				acc.Add(small_table[k], crypto.G)
			}
			(bn256.G1Array(small_table)).MakeAffine() // precompute everything ASAP

			for k := range small_table {
				// convert acc to compressed point and extract last 5 bytes
				//compressed := small_table[k].EncodeCompressed()
				small_table[k].EncodeCompressedToBuf(compressed[:])

				// replace last bytes by j in coded form
				compressed[32] = byte(uint64(j+k) & 0xff)
				compressed[31] = byte((uint64(j+k) >> 8) & 0xff)
				compressed[30] = byte((uint64(j+k) >> 16) & 0xff)

				(t)[i][j+k] = binary.BigEndian.Uint64(compressed[25:])
			}

			if j&8191 == 0 && runtime.GOARCH == "wasm" {

				statusCallback(fmt.Sprintf("%.2f%%", float32(j)*100/float32(len((t)[i]))))

				select {
				case <-ctx.Done():
					return
				case <-readyCn:
					return
				default:
				}
			}
		}

		//fmt.Printf("sorting start\n")
		sort.Sort(t[i])
		//fmt.Printf("sortingcomplete\n")
		//bar.Finish()
	}
	//fmt.Printf("lookuptable complete\n")
	t1 := LookupTable(t)

	tableComputedCn <- &t1
}

// convert point to balance
func (t *LookupTable) Lookup(p *bn256.G1, ctx context.Context, statusCallback func(string)) (uint64, error) {

	// now this big part must be searched in the precomputation lookup table

	//fmt.Printf("decoding balance now\n",)
	var acc bn256.G1

	work_per_loop := new(bn256.G1)

	balance_part := uint64(0)

	balance_per_loop := uint64(len((*t)[0]) * len(*t))

	_ = balance_per_loop

	pcopy := new(bn256.G1).Set(p)

	work_per_loop.ScalarMult(crypto.G, new(big.Int).SetUint64(balance_per_loop))
	work_per_loop = new(bn256.G1).Neg(work_per_loop)

	loop_counter := 0

	balance := uint64(0)

	//  fmt.Printf("jumping into loop %d\n", loop_counter)
	for { // it is an infinite loop

		if loop_counter&2047 == 0 && runtime.GOARCH == "wasm" {
			select {
			case <-ctx.Done():
				return 0, errors.New("Scanning Suspended")
			default:
			}
			statusCallback(strconv.FormatUint(balance, 10))
		}

		if loop_counter != 0 {
			pcopy = new(bn256.G1).Add(pcopy, work_per_loop)
		}
		loop_counter++

		compressed := pcopy.EncodeCompressed()

		compressed[32] = 0
		compressed[31] = 0
		compressed[30] = 0

		big_part := binary.BigEndian.Uint64(compressed[25:])

		for i := range *t {
			index := sort.Search(len((*t)[i]), func(j int) bool { return ((*t)[i][j] & 0xffffffffff000000) >= big_part })
		check_again:
			if index < len((*t)[i]) && ((*t)[i][index]&0xffffffffff000000) == big_part {

				balance_part = ((*t)[i][index]) & 0xffffff
				acc.ScalarMult(crypto.G, new(big.Int).SetUint64(balance+balance_part))

				//if bytes.Equal(acc.EncodeUncompressed(), p.EncodeUncompressed()) { // since we may have part collisions, make sure full point is checked
				if acc.String() == p.String() { // since we may have part collisions, make sure full point is checked
					balance += balance_part
					// fmt.Printf("balance found  %d part(%d) index %d   big part %x\n",balance,balance_part, index, big_part );

					return balance, nil
				}

				// we have failed since it was partial collision, make sure that we can try if possible, another collision
				// this code can be removed if no duplications exist in the first 5 bytes, but a probablity exists for the same
				index++
				goto check_again

			} else {
				// x is not present in data,
				// but i is the index where it would be inserted.

				balance += uint64(len((*t)[i]))
			}

		}

		// from the point we must decrease balance per loop

	}

}
