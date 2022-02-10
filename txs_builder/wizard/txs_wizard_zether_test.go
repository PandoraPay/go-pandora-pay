package wizard

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"testing"
)

func getNewBalance(addr *addresses.Address, amount uint64) *crypto.ElGamal {
	var acckey crypto.Point
	if err := acckey.DecodeCompressed(addr.PublicKey); err != nil {
		panic(err)
	}
	balance := crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
	if amount > 0 {
		balance = balance.Plus(new(big.Int).SetUint64(amount))
	}
	return balance
}

func TestCreateZetherTx(t *testing.T) {

	senderPrivateKey := addresses.GenerateNewPrivateKey()
	senderAdress, err := senderPrivateKey.GenerateAddress(true, nil, 0, nil)
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

	publicKeyIndexes := make(map[string]*WizardZetherPublicKeyIndex)
	publicKeyIndexes[string(senderAdress.PublicKey)] = &WizardZetherPublicKeyIndex{false, 0, senderAdress.Registration}

	fees := make([]*WizardTransactionFee, count)

	transfers := make([]*WizardZetherTransfer, count)
	for i := range transfers {

		dstPrivateKey := addresses.GenerateNewPrivateKey()
		dstAddress, _ := dstPrivateKey.GenerateAddress(true, nil, 0, nil)

		publicKeyIndexes[string(dstAddress.PublicKey)] = &WizardZetherPublicKeyIndex{false, 0, dstAddress.Registration}

		transfers[i] = &WizardZetherTransfer{
			Asset:                config_coins.NATIVE_ASSET_FULL,
			From:                 senderPrivateKey.Key,
			FromDecryptedBalance: amount,
			Destination:          dstAddress.EncodeAddr(),
			Amount:               diff,
			Burn:                 0,
			Data:                 &WizardTransactionData{[]byte{}, false},
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
			ringMemberAddress, _ := ringMemberPrivateKey.GenerateAddress(true, nil, 0, nil)

			publicKeyIndexes[string(ringMemberAddress.PublicKey)] = &WizardZetherPublicKeyIndex{false, 0, ringMemberAddress.Registration}

			ringMemberPoint, _ := ringMemberAddress.GetPoint()
			rings[i][j] = ringMemberPoint.G1()
			emap[config_coins.NATIVE_ASSET_FULL_STRING][ringMemberPoint.G1().String()] = getNewBalance(ringMemberAddress, 0).Serialize()
		}

		fees[i] = &WizardTransactionFee{0, 0, 0, false}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hash := helpers.RandomBytes(32)
	tx, err := CreateZetherTx(transfers, emap, rings, 0, hash, publicKeyIndexes, fees, true, ctx, func(status string) {})
	assert.NoError(t, err)
	assert.NotNil(t, t, tx)

	serialized := tx.SerializeManualToBytes()

	tx2 := &transaction.Transaction{}
	assert.NoError(t, tx2.Deserialize(helpers.NewBufferReader(serialized)))
	assert.NotNil(t, t, tx2)

	assert.NoError(t, tx2.BloomAll())

	assert.Equal(t, true, bytes.Equal(tx.HashManual(), tx2.HashManual()))
	assert.Equal(t, true, bytes.Equal(tx.SerializeForSigning(), tx2.SerializeForSigning()))

	//fmt.Println(hex.EncodeToString(tx.SerializeManualToBytes()))
	//fmt.Println(hex.EncodeToString(tx2.SerializeManualToBytes()))
	assert.Equal(t, true, bytes.Equal(serialized, tx2.SerializeManualToBytes()))

	tx1Base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
	tx2Base := tx2.TransactionBaseInterface.(*transaction_zether.TransactionZether)
	for i, payload := range tx1Base.Payloads {
		for j, publicKey := range payload.Statement.Publickeylist {
			if bytes.Equal(publicKey.EncodeCompressed(), senderAdress.PublicKey) {
				tx2Base.Payloads[i].Statement.CLn[j] = payload.Statement.CLn[j]
				tx2Base.Payloads[i].Statement.CRn[j] = payload.Statement.CRn[j]
			}
			assert.Equal(t, true, bytes.Equal(payload.Statement.CLn[j].EncodeCompressed(), tx2Base.Payloads[i].Statement.CLn[j].EncodeCompressed()))
			assert.Equal(t, true, bytes.Equal(payload.Statement.CRn[j].EncodeCompressed(), tx2Base.Payloads[i].Statement.CRn[j].EncodeCompressed()))
		}
	}

	bytes1, err := json.Marshal(tx)
	assert.NoError(t, err)

	bytes2, err := json.Marshal(tx2)
	assert.NoError(t, err)

	assert.Equal(t, true, bytes.Equal(bytes1, bytes2))

	//let's verify
	assert.Equal(t, true, tx.VerifySignatureManually())
	assert.Equal(t, true, tx2.VerifySignatureManually())

}
