package transaction_zether_payload

import (
	"bytes"
	"errors"
	"fmt"
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

	Parity    bool
	Statement *crypto.Statement // note statement containts fee

	WhisperSender    []byte
	WhisperRecipient []byte

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

	if plainAcc == nil {
		return errors.New("Plain account doesn't exist")
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

	accs, acc, err := dataStorage.GetOrCreateAccount(assetId, plainAcc.AssetFeeLiquidities.Collector, blockHeight, true)
	if err != nil {
		return
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

	if err = payload.Registrations.RegisterNow(payload.Asset, dataStorage, publicKeyList, blockHeight); err != nil {
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
		if acc, err = accs.GetAccount(publicKey, blockHeight); err != nil {
			return
		}

		if acc == nil {
			return errors.New("Private Account doesn't exist")
		}

		balance = acc.GetBalance()
		echanges := crypto.ConstructElGamal(payload.Statement.C[i], payload.Statement.D)
		balance = balance.Add(echanges) // homomorphic addition of changes

		if (i%2 == 0) == payload.Parity { //sender
			//verify sender
			if payload.Statement.CLn[i].String() != balance.Left.String() || payload.Statement.CRn[i].String() != balance.Right.String() {
				return fmt.Errorf("CLn or CRn is not matching for %d", i)
			}
		}

		/**
		STAKING will not update any account
		REWARD will not update any sender account
		*/
		if payload.PayloadScript != SCRIPT_STAKING {

			//Recipient, in case it is delegated it must be a pending stake
			update := false
			if (i%2 == 0) == payload.Parity { //sender
				if payload.PayloadScript != SCRIPT_STAKING_REWARD {
					acc.Balance.Amount = balance
					update = true
				}
			} else { //recipient
				if bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) && acc.DelegatedStake.HasDelegatedStake() {
					if err = acc.DelegatedStake.AddStakePendingStake(echanges, blockHeight); err != nil {
						return
					}
					update = true
				} else {
					acc.Balance.Amount = balance
					update = true
				}
			}

			if update {
				if err = accs.Update(string(publicKey), acc); err != nil {
					return
				}
			}

		}

	}

	if payload.Extra != nil {
		if err = payload.Extra.AfterIncludeTxPayload(txHash, payload.Registrations, payloadIndex, payload.Asset, payload.BurnValue, payload.Statement, publicKeyList, blockHeight, dataStorage); err != nil {
			return
		}
	}

	return
}

func (payload *TransactionZetherPayload) ComputeAllKeys(out map[string]bool, publicKeyList [][]byte) {

	for _, publicKey := range publicKeyList {
		out[string(publicKey)] = true
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

	if !crypto.IsPowerOf2(payload.Statement.RingSize) {
		return fmt.Errorf("corrupted key pointers")
	}

	// check duplicate ring members within the tx
	key_map := map[string]bool{}
	for i := 0; i < payload.Statement.RingSize; i++ {
		key_map[string(payload.Statement.Publickeylist[i].EncodeCompressed())] = true
	}
	if len(key_map) != payload.Statement.RingSize {
		return fmt.Errorf("Duplicated ring members")
	}

	switch payload.PayloadScript {
	case SCRIPT_TRANSFER:
	case SCRIPT_STAKING, SCRIPT_STAKING_REWARD, SCRIPT_ASSET_CREATE, SCRIPT_ASSET_SUPPLY_INCREASE:
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
		w.WriteVariableBytes(payload.Data)
	} else if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED { //fixed 145
		w.Write(payload.Data)
	}

	payload.Statement.SerializeRingSize(w)
	w.WriteBool(payload.Parity)

	payload.Registrations.Serialize(w)

	payload.Statement.Serialize(w, payload.Registrations.Registrations, payload.Parity)

	w.Write(payload.WhisperSender)
	w.Write(payload.WhisperRecipient)

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
	case SCRIPT_STAKING:
		payload.Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraStaking{}
	case SCRIPT_STAKING_REWARD:
		payload.Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraStakingReward{}
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
		if payload.Data, err = r.ReadVariableBytes(config.TRANSACTIONS_MAX_DATA_LENGTH); err != nil {
			return
		}
	case transaction_data.TX_DATA_ENCRYPTED:
		if payload.Data, err = r.ReadBytes(PAYLOAD_LIMIT); err != nil {
			return
		}
	default:
		return errors.New("Invalid Tx.DataVersion")
	}

	payload.Statement = &crypto.Statement{}

	ringPower, ringSize, err := payload.Statement.DeserializeRingSize(r)
	if err != nil {
		return
	}

	if payload.Parity, err = r.ReadBool(); err != nil {
		return
	}

	payload.Registrations = &transaction_zether_registrations.TransactionZetherDataRegistrations{}
	if err = payload.Registrations.Deserialize(r, ringSize); err != nil {
		return
	}

	if err = payload.Statement.Deserialize(r, payload.Registrations.Registrations, payload.Parity); err != nil {
		return
	}

	if payload.WhisperSender, err = r.ReadBytes(32); err != nil {
		return
	}
	if payload.WhisperRecipient, err = r.ReadBytes(32); err != nil {
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

	payload.Proof = &crypto.Proof{}
	if err = payload.Proof.Deserialize(r, int(ringPower)); err != nil {
		return
	}

	if payload.Extra != nil {
		if err = payload.Extra.Deserialize(r); err != nil {
			return
		}
		if err = payload.Extra.UpdateStatement(payload.Statement); err != nil {
			return
		}
	}

	return
}
