package wizard

import (
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	mathrand "math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_zether "pandora-pay/blockchain/transactions/transaction/transaction-zether"
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

	senderPrivateKey := addresses.GenerateNewPrivateKey()
	senderAdress, err := senderPrivateKey.GenerateAddress(true, 0, nil)
	assert.NoError(t, err)

	var amount uint64
	for amount < 1000 {
		amount = mathrand.Uint64()
	}

	count := 5
	emap := make(map[string]map[string][]byte)
	rings := make([][]*bn256.G1, count)

	emap[config.NATIVE_TOKEN_FULL_STRING] = make(map[string][]byte)

	senderPoint, _ := senderAdress.GetPoint()
	emap[config.NATIVE_TOKEN_FULL_STRING][senderPoint.G1().String()] = getNewBalance(senderAdress, amount).Serialize()

	diff := amount / uint64(count)

	publicKeyIndexes := make(map[string]*ZetherPublicKeyIndex)
	publicKeyIndexes[string(senderAdress.PublicKey)] = &ZetherPublicKeyIndex{false, 0, senderAdress.Registration}

	transfers := make([]*ZetherTransfer, 5)
	for i := range transfers {

		dstPrivateKey := addresses.GenerateNewPrivateKey()
		dstAddress, _ := dstPrivateKey.GenerateAddress(true, 0, nil)

		publicKeyIndexes[string(dstAddress.PublicKey)] = &ZetherPublicKeyIndex{false, 0, dstAddress.Registration}

		transfers[i] = &ZetherTransfer{
			Token:              config.NATIVE_TOKEN_FULL,
			From:               senderPrivateKey.Key,
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

		rings[i][0] = senderPoint.G1()

		dstPoint, _ := dstAddress.GetPoint()
		rings[i][1] = dstPoint.G1()
		emap[config.NATIVE_TOKEN_FULL_STRING][dstPoint.G1().String()] = getNewBalance(dstAddress, 0).Serialize()

		for j := 2; j < ringSize; j++ {
			decoyPrivateKey := addresses.GenerateNewPrivateKey()
			decoyAddress, _ := decoyPrivateKey.GenerateAddress(true, 0, nil)

			publicKeyIndexes[string(decoyAddress.PublicKey)] = &ZetherPublicKeyIndex{false, 0, decoyAddress.Registration}

			decoyPoint, _ := decoyAddress.GetPoint()
			rings[i][j] = decoyPoint.G1()
			emap[config.NATIVE_TOKEN_FULL_STRING][decoyPoint.G1().String()] = getNewBalance(decoyAddress, 0).Serialize()
		}
	}

	hash := helpers.RandomBytes(32)
	tx, err := CreateZetherTx(transfers, emap, rings, 0, hash, publicKeyIndexes, func(status string) {})
	assert.NoError(t, err)
	assert.NotNil(t, t, tx)

	serialized := tx.SerializeManualToBytes()

	tx2 := &transaction.Transaction{}
	err = tx2.Deserialize(helpers.NewBufferReader(serialized))
	assert.NoError(t, err)
	assert.NotNil(t, t, tx2)

	//let's fill manually the bloomed data
	for t, payload := range tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).Payloads {
		payload2 := tx2.TransactionBaseInterface.(*transaction_zether.TransactionZether).Payloads[t]

		payload2.Statement.CLn = make([]*bn256.G1, payload.Statement.RingSize)
		payload2.Statement.CRn = make([]*bn256.G1, payload.Statement.RingSize)
		payload2.Statement.Publickeylist = make([]*bn256.G1, payload.Statement.RingSize)

		for i := range payload.Statement.PublicKeysIndexes {
			payload2.Statement.CLn[i] = payload.Statement.CLn[i]
			payload2.Statement.CRn[i] = payload.Statement.CRn[i]
			payload2.Statement.Publickeylist[i] = payload.Statement.Publickeylist[i]
		}

	}

	//fmt.Println("test")
	//fmt.Println(hex.EncodeToString(tx.SerializeManualToBytes()))
	//fmt.Println(hex.EncodeToString(tx2.SerializeManualToBytes()))
	assert.Equal(t, serialized, tx2.SerializeManualToBytes())

	//let's verify
	assert.Equal(t, true, tx.VerifySignatureManually())
	assert.Equal(t, true, tx2.VerifySignatureManually())

}
