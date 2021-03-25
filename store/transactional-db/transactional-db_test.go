package transactional_db

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestCreateTransactionalDB(t *testing.T) {

	db := CreateTransactionalDB()
	tx := db.Update()
	assert.NoError(t, tx.Put([]byte("A"), []byte{1}))
	assert.NoError(t, tx.Put([]byte("B"), []byte{2}))
	assert.NoError(t, tx.Put([]byte("B"), []byte{3}))
	assert.NoError(t, tx.Put([]byte("C"), []byte{4}))

	assert.Equal(t, tx.Get([]byte("C")), []byte{4})
	assert.Equal(t, tx.Get([]byte("B")), []byte{3})

	tx.Commit()
	assert.Equal(t, tx.Get([]byte("C")), []byte{4})
	assert.Equal(t, tx.Get([]byte("B")), []byte{3})
	assert.NoError(t, tx.Store())

	tx2 := db.Update()
	assert.Equal(t, tx2.Get([]byte("B")), []byte{3})
	assert.NoError(t, tx2.Put([]byte("B"), []byte{5}))

	tx3 := db.View()
	assert.Equal(t, tx3.Get([]byte("B")), []byte{3})

	tx2.Commit()
	assert.NoError(t, tx2.Store())
	assert.Equal(t, tx3.Get([]byte("B")), []byte{3})
	tx3.Close()

	tx4 := db.View()
	assert.Equal(t, tx4.Get([]byte("B")), []byte{5})
	tx4.Close()

	time.Sleep(1 * time.Second)

	tx5 := db.View()
	assert.Equal(t, tx5.Get([]byte("B")), []byte{5})
	tx5.Close()
}

func TestCreateTransactionalDB_concurent(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	db := CreateTransactionalDB()

	for i := uint64(0); i < 10000; i++ {

		go func(i uint64) {

			index := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(index, i)
			index = index[:n]

			buf := make([]byte, binary.MaxVarintLen64)

			var tx *TransactionDB
			if rand.Intn(2) == 0 {
				tx = db.View()

				n := binary.PutUvarint(buf, 0)
				outFirst := tx.Get(buf[:n])

				for j := uint64(1); j < 10000; j++ {
					n := binary.PutUvarint(buf, j)
					out := tx.Get(buf[:n])
					assert.Equal(t, out, outFirst)
				}

				tx.Close()
			} else {
				tx = db.Update()
				for j := uint64(0); j < 10000; j++ {
					n := binary.PutUvarint(buf, j)
					assert.NoError(t, tx.Put(buf[:n], index))
					tx.Commit()
				}

				assert.NoError(t, tx.Store())
				tx.Close()
			}
		}(i)

		if i%100 == 0 {
			time.Sleep(100 * time.Millisecond)
		}

	}

}
