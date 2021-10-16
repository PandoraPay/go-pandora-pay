package wizard

import (
	"context"
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_coins"
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

	senderPrivateKey := addresses.GenerateNewPrivateKey()
	senderAdress, err := senderPrivateKey.GenerateAddress(true, 0, nil)
	assert.NoError(t, err)

	var amount uint64
	for amount < 1000 {
		amount = rand.Uint64()
	}

	count := 5
	emap := make(map[string]map[string][]byte)
	rings := make([][]*bn256.G1, count)

	emap[config_coins.NATIVE_ASSET_FULL_STRING] = make(map[string][]byte)

	senderPoint, _ := senderAdress.GetPoint()
	emap[config_coins.NATIVE_ASSET_FULL_STRING][senderPoint.G1().String()] = getNewBalance(senderAdress, amount).Serialize()

	diff := amount / uint64(count)

	publicKeyIndexes := make(map[string]*ZetherPublicKeyIndex)
	publicKeyIndexes[string(senderAdress.PublicKey)] = &ZetherPublicKeyIndex{false, 0, senderAdress.Registration}

	fees := make([]*TransactionsWizardFee, count)

	transfers := make([]*ZetherTransfer, count)
	for i := range transfers {

		dstPrivateKey := addresses.GenerateNewPrivateKey()
		dstAddress, _ := dstPrivateKey.GenerateAddress(true, 0, nil)

		publicKeyIndexes[string(dstAddress.PublicKey)] = &ZetherPublicKeyIndex{false, 0, dstAddress.Registration}

		transfers[i] = &ZetherTransfer{
			Asset:              config_coins.NATIVE_ASSET_FULL,
			From:               senderPrivateKey.Key,
			FromBalanceDecoded: amount,
			Destination:        dstAddress.EncodeAddr(),
			Amount:             diff,
			Burn:               0,
			Data:               &TransactionsWizardData{[]byte{}, false},
		}
		amount -= diff

		power := rand.Intn(4)
		power += 2
		ringSize := int(math.Pow(2, float64(power)))

		rings[i] = make([]*bn256.G1, ringSize)

		rings[i][0] = senderPoint.G1()

		dstPoint, _ := dstAddress.GetPoint()
		rings[i][1] = dstPoint.G1()
		emap[config_coins.NATIVE_ASSET_FULL_STRING][dstPoint.G1().String()] = getNewBalance(dstAddress, 0).Serialize()

		for j := 2; j < ringSize; j++ {
			ringMemberPrivateKey := addresses.GenerateNewPrivateKey()
			ringMemberAddress, _ := ringMemberPrivateKey.GenerateAddress(true, 0, nil)

			publicKeyIndexes[string(ringMemberAddress.PublicKey)] = &ZetherPublicKeyIndex{false, 0, ringMemberAddress.Registration}

			ringMemberPoint, _ := ringMemberAddress.GetPoint()
			rings[i][j] = ringMemberPoint.G1()
			emap[config_coins.NATIVE_ASSET_FULL_STRING][ringMemberPoint.G1().String()] = getNewBalance(ringMemberAddress, 0).Serialize()
		}

		fees[i] = &TransactionsWizardFee{0, 0, 0, false}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hash := helpers.RandomBytes(32)
	tx, err := CreateZetherTx(transfers, emap, rings, 0, hash, publicKeyIndexes, fees, true, ctx, func(status string) {})
	assert.NoError(t, err)
	assert.NotNil(t, t, tx)

	serialized := tx.SerializeManualToBytes()

	tx2 := &transaction.Transaction{}
	err = tx2.Deserialize(helpers.NewBufferReader(serialized))
	assert.NoError(t, err)
	assert.NotNil(t, t, tx2)

	//fmt.Println("test")
	//fmt.Println(hex.EncodeToString(tx.SerializeManualToBytes()))
	//fmt.Println(hex.EncodeToString(tx2.SerializeManualToBytes()))
	assert.Equal(t, serialized, tx2.SerializeManualToBytes())

	//let's verify
	assert.Equal(t, true, tx.VerifySignatureManually())
	assert.Equal(t, true, tx2.VerifySignatureManually())

}
