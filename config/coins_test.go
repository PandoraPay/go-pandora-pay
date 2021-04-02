package config

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestConvertToBase(t *testing.T) {

	for i := 0; i < 10000; i++ {
		no := rand.Uint64()

		base := ConvertToBase(no)
		_, err := ConvertToUnits(base)

		assert.NoError(t, err)
	}

}
