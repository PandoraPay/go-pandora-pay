package min_max_heap

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"testing"
	"time"
)

func sum(v []uint64) uint64 {
	out := uint64(0)
	for _, x := range v {
		out += x
	}
	return out
}
func getMax(v []uint64) uint64 {
	max := v[0]
	for _, x := range v {
		if max < x {
			max = x
		}
	}
	return max
}

func getMin(v []uint64) uint64 {
	min := v[0]
	for _, x := range v {
		if min > x {
			min = x
		}
	}
	return min
}

func TestCreateMaxHeapMemory(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	v := []uint64{6, 5, 3, 7, 2, 8}
	keys := make([][]byte, len(v))

	for i := range v {
		keys[i] = helpers.RandomBytes(cryptography.PublicKeySize)
	}

	maxHeap := NewHeapMemory(func(a, b uint64) bool {
		return b < a
	})

	for i := 0; i < len(v); i++ {
		assert.Nil(t, maxHeap.Insert(v[i], keys[i]))
		top, err := maxHeap.GetTop()

		assert.Nil(t, err)
		assert.Equal(t, top.Score, getMax(v[:i+1]), "Max is not matching")
	}

	index := rand.Intn(len(v))
	x := append(v[:index], v[index+1:]...)
	assert.Nil(t, maxHeap.DeleteByKey(keys[index]))

	assert.Equal(t, maxHeap.getSize(), uint64(len(x)))

	final2 := uint64(0)
	for range x {
		el, err := maxHeap.RemoveTop()
		assert.Nil(t, err)
		final2 += el.Score
	}
	assert.Equal(t, sum(x), final2)
}

func TestCreateMaxHeapMemory2(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	v := []uint64{}
	keys := make([][]byte, len(v))

	dict := map[string]int{}
	for i := 0; i < 10000; i++ {
		v = append(v, rand.Uint64())
		keys = append(keys, helpers.RandomBytes(cryptography.PublicKeySize))
		dict[string(keys[i])] = i
	}

	maxHeap := NewHeapMemory(func(a, b uint64) bool {
		return b < a
	})
	for i := 0; i < len(v); i++ {
		assert.Nil(t, maxHeap.Insert(v[i], keys[i]))
		top, err := maxHeap.GetTop()

		assert.Nil(t, err)
		assert.Equal(t, top.Score, getMax(v[:i+1]), "Max is not matching")
	}

	index := rand.Intn(len(v))
	x := append(v[:index], v[index+1:]...)
	assert.Nil(t, maxHeap.DeleteByKey(keys[index]))

	assert.Equal(t, maxHeap.getSize(), uint64(len(x)))

	final2 := uint64(0)
	for range x {
		el, err := maxHeap.RemoveTop()
		assert.Nil(t, err)
		final2 += el.Score
	}
	assert.Equal(t, sum(x), final2)
}

func TestCreateMinHeapMemory(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	v := []uint64{}
	keys := make([][]byte, len(v))

	dict := map[string]int{}
	for i := 0; i < 10000; i++ {
		v = append(v, rand.Uint64())
		keys = append(keys, helpers.RandomBytes(cryptography.PublicKeySize))
		dict[string(keys[i])] = i
	}

	minHeap := NewHeapMemory(func(a, b uint64) bool {
		return a < b
	})
	for i := 0; i < len(v); i++ {
		assert.Nil(t, minHeap.Insert(v[i], keys[i]))
		top, err := minHeap.GetTop()

		assert.Nil(t, err)
		assert.Equal(t, top.Score, getMin(v[:i+1]), "Min is not matching")
	}

	index := rand.Intn(len(v))
	x := append(v[:index], v[index+1:]...)
	assert.Nil(t, minHeap.DeleteByKey(keys[index]))

	assert.Equal(t, minHeap.getSize(), uint64(len(x)))

	final2 := uint64(0)
	for range x {
		el, err := minHeap.RemoveTop()
		assert.Nil(t, err)
		final2 += el.Score
	}
	assert.Equal(t, sum(x), final2)
}
