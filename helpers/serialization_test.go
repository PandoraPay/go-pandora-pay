package helpers

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerializeNumber(t *testing.T) {

	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, 0)
	assert.Equal(t, n, 1)
	assert.Equal(t, b[0], uint8(0))

	n = binary.PutUvarint(b, 1)
	assert.Equal(t, n, 1)
	assert.Equal(t, b[0], uint8(1))

	n = binary.PutUvarint(b, 120)
	assert.Equal(t, n, 1)
	assert.Equal(t, b[0], uint8(120))

	n = binary.PutUvarint(b, 126)
	assert.Equal(t, n, 1)
	assert.Equal(t, b[0], uint8(126))

	n = binary.PutUvarint(b, 127)
	assert.Equal(t, n, 1)
	assert.Equal(t, b[0], uint8(127))

	n = binary.PutUvarint(b, 128)
	assert.Equal(t, n, 2)
	assert.Equal(t, b[0], uint8(128))

	a, done := binary.Uvarint(b)
	assert.Equal(t, a, uint64(128))
	assert.Equal(t, done > 0, true)

	binary.PutUvarint(b, 0xFFFFFFFFFFFFFFFF)
	a, done = binary.Uvarint(b)
	assert.Equal(t, a, uint64(0xFFFFFFFFFFFFFFFF))
	assert.Equal(t, done > 0, true)

	binary.PutUvarint(b, 0xFFFFFFFFFFFFFFFC)
	a, done = binary.Uvarint(b)
	assert.Equal(t, a, uint64(0xFFFFFFFFFFFFFFFC))
	assert.Equal(t, done > 0, true)

}

func TestDeserializeNumber(t *testing.T) {

	b := make([]byte, binary.MaxVarintLen64)
	for i := 0; i < 100; i++ {

		no := RandomUint64()

		binary.PutUvarint(b, no)

		a, done := binary.Uvarint(b)
		assert.Equal(t, a, no)
		assert.Equal(t, done > 0, true)

	}

}
