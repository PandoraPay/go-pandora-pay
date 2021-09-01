package wizard

import (
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	mathrand "math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"testing"
)

func getNewBalance(addr *addresses.Address, amount uint64) *crypto.ElGamal {
	point, _ := addr.GetPoint()
	balance := crypto.ConstructElGamal(point.G1(), crypto.ElGamal_BASE_G)
	if amount > 0 {
		balance = balance.Plus(new(big.Int).SetUint64(amount))
	}
	return balance
}

func TestCreateZetherTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	address, err := privateKey.GenerateAddress(false, 0, nil)
	assert.NoError(t, err)

	var amount uint64
	for amount < 1000 {
		amount = mathrand.Uint64()
	}

	count := 5
	emap := make(map[string]map[string][]byte)
	rings := make([][]*bn256.G1, count)

	emap[config.NATIVE_TOKEN_STRING] = make(map[string][]byte)

	point, _ := address.GetPoint()
	emap[config.NATIVE_TOKEN_STRING][point.G1().String()] = getNewBalance(address, amount).Serialize()

	diff := amount / uint64(count)

	transfers := make([]*ZetherTransfer, 5)
	for i := range transfers {

		dstPrivateKey := addresses.GenerateNewPrivateKey()
		dstAddress, _ := dstPrivateKey.GenerateAddress(false, 0, nil)

		transfers[i] = &ZetherTransfer{
			Token:              config.NATIVE_TOKEN,
			From:               privateKey.Key,
			FromBalanceDecoded: amount,
			Destination:        dstAddress.EncodeAddr(),
			Amount:             diff,
			Burn:               0,
			Data:               &TransactionsWizardData{[]byte{}, false},
		}
		amount -= diff

		power := mathrand.Int() % 4
		power += 2
		ringSize := int(math.Pow(2, float64(power)))

		rings[i] = make([]*bn256.G1, ringSize)

		rings[i][0] = point.G1()

		dstPoint, _ := dstAddress.GetPoint()
		rings[i][1] = dstPoint.G1()
		emap[config.NATIVE_TOKEN_STRING][dstPoint.G1().String()] = getNewBalance(dstAddress, 0).Serialize()

		for j := 2; j < ringSize; j++ {
			decoyPrivateKey := addresses.GenerateNewPrivateKey()
			decoyAddress, _ := decoyPrivateKey.GenerateAddress(false, 0, nil)
			decoyPoint, _ := decoyAddress.GetPoint()
			rings[i][j] = decoyPoint.G1()
			emap[config.NATIVE_TOKEN_STRING][decoyPoint.G1().String()] = getNewBalance(decoyAddress, 0).Serialize()
		}
	}

	hash := helpers.RandomBytes(32)
	tx, err := CreateZetherTx(transfers, emap, rings, 0, hash, func(status string) {})
	assert.NoError(t, err)
	assert.NotNil(t, t, tx)

	serialized := tx.SerializeManualToBytes()

	tx2 := &transaction.Transaction{}
	err = tx2.Deserialize(helpers.NewBufferReader(serialized))
	assert.NoError(t, err)
	assert.NotNil(t, t, tx2)

}
