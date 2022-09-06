package min_max_heap

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"math/rand"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"testing"
	"time"
)

func sum(v []float64) float64 {
	out := float64(0)
	for _, x := range v {
		out += x
	}
	return out
}
func getMax(v []float64) float64 {
	max := v[0]
	for _, x := range v {
		if max < x {
			max = x
		}
	}
	return max
}

func getMin(v []float64) float64 {
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

	v := []float64{6, 5, 3, 7, 2, 8}
	keys := make([][]byte, len(v))

	for i := range v {
		keys[i] = helpers.RandomBytes(cryptography.PublicKeySize)
	}

	maxHeap := NewHeapMemory(func(a, b float64) bool {
		return b < a
	})

	for i := 0; i < len(v); i++ {
		assert.Nil(t, maxHeap.Insert(v[i], keys[i]))
		top, err := maxHeap.GetTop()

		assert.Nil(t, err)
		assert.Equal(t, top.Score, getMax(v[:i+1]), "Max is not matching")
	}

	index := rand.Intn(len(v))
	x := slices.Delete(v, index, index+1)
	assert.Nil(t, maxHeap.DeleteByKey(keys[index]))

	assert.Equal(t, maxHeap.GetSize(), uint64(len(x)))

	final2 := float64(0)
	for range x {
		el, err := maxHeap.RemoveTop()
		assert.Nil(t, err)
		final2 += el.Score
	}
	assert.Equal(t, sum(x), final2)
}

func TestCreateMaxHeapMemory2(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	v := []float64{}
	keys := make([][]byte, len(v))

	dict := map[string]int{}
	for i := 0; i < 10000; i++ {
		v = append(v, float64(rand.Uint32()))
		keys = append(keys, helpers.RandomBytes(cryptography.PublicKeySize))
		dict[string(keys[i])] = i
	}

	maxHeap := NewHeapMemory(func(a, b float64) bool {
		return b < a
	})
	for i := 0; i < len(v); i++ {
		assert.Nil(t, maxHeap.Insert(v[i], keys[i]))
		top, err := maxHeap.GetTop()

		assert.Nil(t, err)
		assert.Equal(t, top.Score, getMax(v[:i+1]), "Max is not matching")
	}

	index := rand.Intn(len(v))
	x := slices.Delete(v, index, index+1)
	assert.Nil(t, maxHeap.DeleteByKey(keys[index]))

	assert.Equal(t, maxHeap.GetSize(), uint64(len(x)))

	final2 := float64(0)
	for range x {
		el, err := maxHeap.RemoveTop()
		assert.Nil(t, err)
		final2 += el.Score
	}
	assert.Equal(t, sum(x), final2)

	top, err := maxHeap.GetTop()
	assert.Nil(t, top)
	assert.Nil(t, err)
}

func TestCreateMinHeapMemory(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	v := []float64{}
	keys := make([][]byte, len(v))

	dict := map[string]int{}
	for i := 0; i < 10000; i++ {
		v = append(v, float64(rand.Uint32()))
		keys = append(keys, helpers.RandomBytes(cryptography.PublicKeySize))
		dict[string(keys[i])] = i
	}

	minHeap := NewHeapMemory(func(a, b float64) bool {
		return a < b
	})
	for i := 0; i < len(v); i++ {
		assert.Nil(t, minHeap.Insert(v[i], keys[i]))
		top, err := minHeap.GetTop()

		assert.Nil(t, err)
		assert.Equal(t, top.Score, getMin(v[:i+1]), "Min is not matching")
	}

	index := rand.Intn(len(v))
	x := slices.Delete(v, index, index+1)
	assert.Nil(t, minHeap.DeleteByKey(keys[index]))

	assert.Equal(t, minHeap.GetSize(), uint64(len(x)))

	final2 := float64(0)
	for range x {
		el, err := minHeap.RemoveTop()
		assert.Nil(t, err)
		final2 += el.Score
	}
	assert.Equal(t, sum(x), final2)

	top, err := minHeap.GetTop()
	assert.Nil(t, top)
	assert.Nil(t, err)
}
