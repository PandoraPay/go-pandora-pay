package min_heap

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func getMin(v []uint64) uint64 {
	min := v[0]
	for _, x := range v {
		if min > x {
			min = x
		}
	}
	return min
}

func TestCreateMinHeapMemory(t *testing.T) {
	v := []uint64{6, 5, 3, 7, 2, 8}
	minHeap := CreateMinHeapMemory()
	for i := 0; i < len(v); i++ {
		assert.Nil(t, minHeap.Insert(v[i], nil))
		min, err := minHeap.GetMin()

		assert.Nil(t, err)
		if min.Score != getMin(v[:i+1]) {
			t.Error("Minim is not matching")
		}
	}
}

func TestCreateMinHeapMemory2(t *testing.T) {
	v := []uint64{}
	for i := 0; i < 10000; i++ {
		v = append(v, rand.Uint64())
	}

	minHeap := CreateMinHeapMemory()
	for i := 0; i < len(v); i++ {
		assert.Nil(t, minHeap.Insert(v[i], nil))
		min, err := minHeap.GetMin()

		assert.Nil(t, err)
		if min.Score != getMin(v[:i+1]) {
			t.Error("Minim is not matching")
		}
	}
}
