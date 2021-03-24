package transactional_db

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
	assert.NoError(t, tx.Store())
	assert.Equal(t, tx3.Get([]byte("B")), []byte{3})
}
