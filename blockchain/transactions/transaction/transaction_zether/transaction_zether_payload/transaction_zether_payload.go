package transaction_zether_payload

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config"
	"pandora-pay/config/config_assets"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayload struct {
	PayloadScript PayloadScriptType

	Asset     []byte
	BurnValue uint64

	DataVersion transaction_data.TransactionDataVersion
	Data        []byte // sender position in ring representation in a byte, upto 256 ring
	// 144 byte payload  ( to implement specific functionality such as delivery of keys etc), user dependent encryption

	Registrations *transaction_zether_registrations.TransactionZetherDataRegistrations

	Statement *crypto.Statement // note statement containts fee

	FeeRate         uint64 //serialized only if asset is not native
	FeeLeadingZeros byte

	Proof *crypto.Proof
	Extra transaction_zether_payload_extra.TransactionZetherPayloadExtraInterface
}

func (payload *TransactionZetherPayload) processAssetFee(assetId []byte, txFee, txFeeRate uint64, txFeeLeadingZeros byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	key, err := dataStorage.AstsFeeLiquidityCollection.GetTopLiquidity(assetId)
	if err != nil {
		return err
	}

	if key == nil {
		return errors.New("There is no Asset Fee Liquidity Available")
	}

	plainAcc, err := dataStorage.PlainAccs.GetPlainAccount(key, blockHeight)
	if err != nil {
		return
	}

	assetFeeLiquidity := plainAcc.AssetFeeLiquidities.GetLiquidity(assetId)

	if assetFeeLiquidity.Rate < txFeeRate {
		return errors.New("assetFeeLiquidity.Rate < txFeeRate")
	}

	final := txFee //it will copy
	if err = helpers.SafeUint64Mul(&final, txFeeRate); err != nil {
		return
	}
	final = final / helpers.Pow10(txFeeLeadingZeros)

	if err = dataStorage.SubtractUnclaimed(plainAcc, final, blockHeight); err != nil {
		return
	}

	accs, err := dataStorage.AccsCollection.GetMap(assetId)
	if err != nil {
		return
	}

	acc, err := accs.GetAccount(plainAcc.AssetFeeLiquidities.Collector)
	if err != nil {
		return
	}

	if acc == nil {
		if acc, err = accs.CreateAccount(plainAcc.AssetFeeLiquidities.Collector); err != nil {
			return
		}
	}

	acc.Balance.AddBalanceUint(txFee)
	return accs.Update(string(plainAcc.AssetFeeLiquidities.Collector), acc)
}

func (payload *TransactionZetherPayload) IncludePayload(txHash []byte, payloadIndex byte, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	var acc *account.Account
	var balance *crypto.ElGamal

	if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
		if err = payload.processAssetFee(payload.Asset, payload.Statement.Fee, payload.FeeRate, payload.FeeLeadingZeros, blockHeight, dataStorage); err != nil {
			return
		}
	}

	if err = payload.Registrations.RegisterNow(dataStorage, publicKeyList); err != nil {
		return
	}

	if payload.Extra != nil {
		if err = payload.Extra.BeforeIncludeTxPayload(txHash, payload.Registrations, payloadIndex, payload.Asset, payload.BurnValue, payload.Statement, publicKeyList, blockHeight, dataStorage); err != nil {
			return
		}
	}

	if accs, err = dataStorage.AccsCollection.GetMap(payload.Asset); err != nil {
		return
	}

	if len(payload.Statement.Publickeylist) != len(publicKeyList) {
		return errors.New("publicKeyList was not precomputed")
	}

	for i, publicKey := range publicKeyList {

		if acc, err = accs.GetAccount(publicKey); err != nil {
			return
		}

		if acc == nil {
			if acc, err = accs.CreateAccount(publicKey); err != nil {
				return
			}
		}

		balance = acc.GetBalance()
		echanges := crypto.ConstructElGamal(payload.Statement.C[i], payload.Statement.D)
		balance = balance.Add(echanges) // homomorphic addition of changes

		//verify
		if payload.Statement.CLn[i].String() != balance.Left.String() || payload.Statement.CRn[i].String() != balance.Right.String() {
			return errors.New("CLn or CRn is not matching")
		}

		acc.Balance.Amount = balance
		if err = accs.Update(string(publicKey), acc); err != nil {
			return
		}
	}

	if payload.Extra != nil {
		if err = payload.Extra.IncludeTxPayload(txHash, payload.Registrations, payloadIndex, payload.Asset, payload.BurnValue, payload.Statement, publicKeyList, blockHeight, dataStorage); err != nil {
			return
		}
	}

	return
}

func (payload *TransactionZetherPayload) ComputeAllKeys(out map[string]bool) {

	for _, publicKey := range payload.Statement.Publickeylist {
		out[string(publicKey.EncodeCompressed())] = true
	}

	if payload.Extra != nil {
		payload.Extra.ComputeAllKeys(out)
	}
}

func (payload *TransactionZetherPayload) Validate(payloadIndex byte) (err error) {

	if bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
		if payload.FeeLeadingZeros != 0 || payload.FeeRate != 0 {
			return errors.New(" Leading Zeros must be zero")
		}
	} else {
		if payload.FeeLeadingZeros > config_assets.ASSETS_DECIMAL_SEPARATOR_MAX_BYTE {
			return errors.New("Invalid Leading Zeros")
		}
	}

	// check sanity
	if payload.Statement.RingSize < 2 { // ring size minimum 4
		return fmt.Errorf("RingSize cannot be less than 2")
	}

	if payload.Statement.RingSize > config.TRANSACTIONS_ZETHER_RING_MAX { // ring size current limited to 256
		return fmt.Errorf("RingSize cannot be that big")
	}

	if !crypto.IsPowerOf2(int(payload.Statement.RingSize)) {
		return fmt.Errorf("corrupted key pointers")
	}

	// check duplicate ring members within the tx
	key_map := map[string]bool{}
	for i := 0; i < int(payload.Statement.RingSize); i++ {
		key_map[string(payload.Statement.Publickeylist[i].EncodeCompressed())] = true
	}
	if len(key_map) != int(payload.Statement.RingSize) {
		return fmt.Errorf("Duplicated ring members")
	}

	switch payload.PayloadScript {
	case SCRIPT_TRANSFER:
	case SCRIPT_DELEGATE_STAKE, SCRIPT_CLAIM, SCRIPT_ASSET_CREATE, SCRIPT_ASSET_SUPPLY_INCREASE:
		if payload.Extra == nil {
			return errors.New("extra is not assigned")
		}
		if err = payload.Extra.Validate(payload.Registrations, payloadIndex, payload.Asset, payload.BurnValue, payload.Statement); err != nil {
			return
		}
	default:
		return errors.New("Invalid Zether PayloadScript")
	}

	return
}

func (payload *TransactionZetherPayload) Serialize(w *helpers.BufferWriter, inclSignature bool) {

	w.WriteUvarint(uint64(payload.PayloadScript))

	w.WriteAsset(payload.Asset)
	w.WriteUvarint(payload.BurnValue)

	w.WriteByte(byte(payload.DataVersion))
	if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT { //variable
		w.WriteUvarint(uint64(len(payload.Data)))
		w.Write(payload.Data)
	} else if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED { //fixed 145
		w.Write(payload.Data)
	}

	payload.Registrations.Serialize(w)

	payload.Statement.Serialize(w)

	if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
		w.WriteUvarint(payload.FeeRate)
		w.WriteByte(payload.FeeLeadingZeros)
	}

	if inclSignature {
		payload.Proof.Serialize(w)
	}

	if payload.Extra != nil {
		payload.Extra.Serialize(w, inclSignature)
	}
}

func (payload *TransactionZetherPayload) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	payload.PayloadScript = PayloadScriptType(n)
	switch payload.PayloadScript {
	case SCRIPT_TRANSFER:
		payload.Extra = nil
	case SCRIPT_DELEGATE_STAKE:
		payload.Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake{}
	case SCRIPT_CLAIM:
		payload.Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraClaim{}
	case SCRIPT_ASSET_CREATE:
		payload.Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate{}
	case SCRIPT_ASSET_SUPPLY_INCREASE:
		payload.Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease{}
	default:
		return errors.New("INVALID SCRIPT TYPE")
	}

	if payload.Asset, err = r.ReadAsset(); err != nil {
		return
	}
	if payload.BurnValue, err = r.ReadUvarint(); err != nil {
		return
	}

	var dataVersion byte
	if dataVersion, err = r.ReadByte(); err != nil {
		return
	}

	payload.DataVersion = transaction_data.TransactionDataVersion(dataVersion)

	switch payload.DataVersion {
	case transaction_data.TX_DATA_NONE:
	case transaction_data.TX_DATA_PLAIN_TEXT:
		if n, err = r.ReadUvarint(); err != nil {
			return
		}
		if n == 0 || n > config.TRANSACTIONS_MAX_DATA_LENGTH {
			return errors.New("Tx.Data length is invalid")
		}
		if payload.Data, err = r.ReadBytes(int(n)); err != nil {
			return
		}
	case transaction_data.TX_DATA_ENCRYPTED:
		if payload.Data, err = r.ReadBytes(PAYLOAD_LIMIT); err != nil {
			return
		}
	default:
		return errors.New("Invalid Tx.DataVersion")
	}

	payload.Registrations = &transaction_zether_registrations.TransactionZetherDataRegistrations{}
	if err = payload.Registrations.Deserialize(r); err != nil {
		return
	}

	payload.Statement = &crypto.Statement{}
	if err = payload.Statement.Deserialize(r); err != nil {
		return
	}

	if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
		if payload.FeeRate, err = r.ReadUvarint(); err != nil {
			return
		}
		if payload.FeeLeadingZeros, err = r.ReadByte(); err != nil {
			return
		}
	}

	m := int(math.Log2(float64(payload.Statement.RingSize)))
	if math.Pow(2, float64(m)) != float64(payload.Statement.RingSize) {
		return errors.New("log failed")
	}

	payload.Proof = &crypto.Proof{}
	if err = payload.Proof.Deserialize(r, m); err != nil {
		return
	}

	if payload.Extra != nil {
		return payload.Extra.Deserialize(r)
	}

	return
}
