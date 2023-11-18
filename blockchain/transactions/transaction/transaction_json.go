package transaction

import (
	"encoding/json"
	"errors"
	msgpack "github.com/vmihailenco/msgpack/v5"
	"math"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations/transaction_zether_registration"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/helpers/advanced_buffers"
)

type json_TransactionDataRegistration struct {
	RegistrationType           transaction_zether_registration.TransactionZetherDataRegistrationType `json:"registrationType"  msgpack:"registrationType"`
	RegistrationStaked         bool                                                                  `json:"registrationStaked" msgpack:"registrationStaked"`
	RegistrationSpendPublicKey []byte                                                                `json:"registrationSpendPublicKey" msgpack:"registrationSpendPublicKey"`
	RegistrationSignature      []byte
}

type Json_Transaction struct {
	Version    transaction_type.TransactionVersion `json:"version" msgpack:"version"`
	Size       uint64                              `json:"size" msgpack:"size"`
	SpaceExtra uint64                              `json:"spaceExtra" msgpack:"spaceExtra"`
	Hash       []byte                              `json:"hash" msgpack:"hash"`
}

type json_TransactionSimple struct {
	*Json_Transaction
	TxScript    transaction_simple.ScriptType           `json:"txScript" msgpack:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion" msgpack:"dataVersion"`
	Data        []byte                                  `json:"data" msgpack:"data"`
	Nonce       uint64                                  `json:"nonce" msgpack:"nonce"`
	Fee         uint64                                  `json:"fee" msgpack:"fee"`
	Vin         *json_TransactionSimpleInput            `json:"vin" msgpack:"vin"`
	Extra       interface{}                             `json:"extra" msgpack:"extra"`
}

type json_TransactionSimpleInput struct {
	PublicKey []byte `json:"publicKey,omitempty" msgpack:"publicKey,omitempty"` //32
	Signature []byte `json:"signature" msgpack:"signature"`                     //64
}

type json_TransactionSimpleNothing struct {
	*json_TransactionSimple
}

type json_Only_TransactionSimpleExtraUpdateAssetFeeLiquidity struct {
	Liquidities  []*asset_fee_liquidity.AssetFeeLiquidity `json:"liquidities"`
	NewCollector bool                                     `json:"newCollector"`
	Collector    []byte                                   `json:"collector"`
}

type json_Only_TransactionSimpleExtraResolutionConditionalPayment struct {
	TxId               []byte   `json:"txId"`
	PayloadIndex       byte     `json:"payloadIndex"`
	Resolution         bool     `json:"resolution"`
	MultisigPublicKeys [][]byte `json:"multisigPublicKeys"`
	Signatures         [][]byte `json:"signatures"`
}

type json_Only_TransactionZether struct {
	ChainHeight     uint64                          `json:"chainHeight"  msgpack:"chainHeight"`
	ChainKernelHash []byte                          `json:"chainKernelHash"  msgpack:"chainKernelHash"`
	Payloads        []*json_Only_TransactionPayload `json:"payloads"  msgpack:"payloads"`
}

type json_Only_TransactionZetherPayloadExtraStaking struct {
}

type json_Only_TransactionZetherPayloadExtraStakingReward struct {
	Reward                            uint64 `json:"reward"  msgpack:"reward"`
	TemporaryAccountRegistrationIndex uint64 `json:"temporaryAccountRegistrationIndex"  msgpack:"temporaryAccountRegistrationIndex"`
}

type json_Only_TransactionZetherPayloadExtraSpend struct {
	SenderSpendPublicKey []byte `json:"senderSpendPublicKey"  msgpack:"senderSpendPublicKey"`
	SenderSpendSignature []byte `json:"senderSpendSignature"  msgpack:"senderSpendSignature"`
}

type json_Only_TransactionZetherPayloadExtraAssetCreate struct {
	Asset *asset.Asset `json:"asset"  msgpack:"asset"`
}

type json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease struct {
	AssetId              []byte `json:"assetId"  msgpack:"assetId"`
	ReceiverPublicKey    []byte `json:"receiverPublicKey"  msgpack:"receiverPublicKey"` //must be registered before
	Value                uint64 `json:"value"  msgpack:"value"`
	AssetSupplyPublicKey []byte `json:"assetSupplyPublicKey"  msgpack:"assetSupplyPublicKey"` //TODO: it can be bloomed
	AssetSignature       []byte `json:"assetSignature"  msgpack:"assetSignature"`
}

type json_Only_TransactionZetherPayloadExtraPlainAccountFund struct {
	PlainAccountPublicKey []byte `json:"plainAccountPublicKey"  msgpack:"plainAccountPublicKey"`
}

type json_Only_TransactionZetherPayloadExtraConditionalPayment struct {
	Deadline           uint64   `json:"deadline" msgpack:"deadline"`
	DefaultResolution  bool     `json:"defaultResolution" msgpack:"defaultResolution"`
	MultisigThreshold  byte     `json:"multisigThreshold" msgpack:"multisigThreshold"`
	MultisigPublicKeys [][]byte `json:"multisigPublicKeys" msgpack:"multisigPublicKeys"`
}

type json_Only_TransactionZetherStatement struct {
	RingSize      int      `json:"ringSize"  msgpack:"ringSize"`
	CLn           [][]byte `json:"cLn"  msgpack:"cLn"`
	CRn           [][]byte `json:"cRn"  msgpack:"cRn"`
	Publickeylist [][]byte `json:"publickeylist"  msgpack:"publickeylist"`
	C             [][]byte `json:"c"  msgpack:"c"`
	D             []byte   `json:"d"  msgpack:"d"`
	Fee           uint64   `json:"fee"  msgpack:"fee"`
}

type json_Only_TransactionPayload struct {
	PayloadScript    transaction_zether_payload_script.PayloadScriptType `json:"payloadScript"  msgpack:"payloadScript"`
	Asset            []byte                                              `json:"asset"  msgpack:"asset"`
	BurnValue        uint64                                              `json:"burnValue"  msgpack:"burnValue"`
	DataVersion      transaction_data.TransactionDataVersion             `json:"dataVersion"  msgpack:"dataVersion"`
	Data             []byte                                              `json:"data"  msgpack:"data"`
	Registrations    []*json_TransactionDataRegistration                 `json:"registrations"  msgpack:"registrations"`
	Parity           bool                                                `json:"parity" msgpack:"parity"`
	Statement        *json_Only_TransactionZetherStatement               `json:"statement"  msgpack:"statement"`
	WhisperSender    []byte                                              `json:"whisperSender" msgpack:"whisperSender"`
	WhisperRecipient []byte                                              `json:"whisperRecipient" msgpack:"whisperRecipient"`
	FeeRate          uint64                                              `json:"feeRate"  msgpack:"feeRate"`
	FeeLeadingZeros  byte                                                `json:"feeLeadingZeros"  msgpack:"feeLeadingZeros"`
	Proof            []byte                                              `json:"proof"  msgpack:"proof"`
	Extra            interface{}                                         `json:"extra"  msgpack:"extra"`
}

type json_TransactionZether struct {
	*Json_Transaction
	*json_Only_TransactionZether
}

func marshalJSON(tx *Transaction, marshal func(any) ([]byte, error)) ([]byte, error) {

	txJson := &Json_Transaction{
		tx.Version,
		tx.Bloom.Size,
		tx.SpaceExtra,
		tx.Bloom.Hash,
	}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		var vinJson *json_TransactionSimpleInput

		if base.HasVin() {
			vinJson = &json_TransactionSimpleInput{
				base.Vin.PublicKey,
				base.Vin.Signature,
			}
		}

		simpleJson := &json_TransactionSimple{
			txJson,
			base.TxScript,
			base.DataVersion,
			base.Data,
			base.Nonce,
			base.Fee,
			vinJson,
			nil,
		}

		switch base.TxScript {
		case transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity)
			simpleJson.Extra = json_Only_TransactionSimpleExtraUpdateAssetFeeLiquidity{
				extra.Liquidities,
				extra.NewCollector,
				extra.Collector,
			}
		case transaction_simple.SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment)
			simpleJson.Extra = json_Only_TransactionSimpleExtraResolutionConditionalPayment{
				extra.TxId,
				extra.PayloadIndex,
				extra.Resolution,
				extra.MultisigPublicKeys,
				extra.Signatures,
			}
		case transaction_simple.SCRIPT_SIMPLE_NOTHING:
		case transaction_simple
		default:
			return nil, errors.New("Invalid simple.TxScript")
		}

		return marshal(simpleJson)

	case transaction_type.TX_ZETHER:
		base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

		payloadsJson := make([]*json_Only_TransactionPayload, len(base.Payloads))
		for i, payload := range base.Payloads {

			registrations := make([]*json_TransactionDataRegistration, len(payload.Registrations.Registrations))
			for i, reg := range payload.Registrations.Registrations {
				if reg != nil {
					registrations[i] = &json_TransactionDataRegistration{
						reg.RegistrationType,
						reg.RegistrationStaked,
						reg.RegistrationSpendPublicKey,
						reg.RegistrationSignature,
					}
				}
			}

			statementJson := &json_Only_TransactionZetherStatement{
				payload.Statement.RingSize,
				helpers.ConvertBN256Array(payload.Statement.CLn),
				helpers.ConvertBN256Array(payload.Statement.CRn),
				helpers.ConvertBN256Array(payload.Statement.Publickeylist),
				helpers.ConvertBN256Array(payload.Statement.C),
				payload.Statement.D.EncodeCompressed(),
				payload.Statement.Fee,
			}

			w := advanced_buffers.NewBufferWriter()
			payload.Proof.Serialize(w)
			proofJson := w.Bytes()

			var extra interface{}

			switch payload.PayloadScript {
			case transaction_zether_payload_script.SCRIPT_TRANSFER:
				//no payload
			case transaction_zether_payload_script.SCRIPT_STAKING:
				//payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraStaking)
				extra = &json_Only_TransactionZetherPayloadExtraStaking{}
			case transaction_zether_payload_script.SCRIPT_STAKING_REWARD:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraStakingReward)
				extra = &json_Only_TransactionZetherPayloadExtraStakingReward{
					payloadExtra.Reward,
					payloadExtra.TemporaryAccountRegistrationIndex,
				}
			case transaction_zether_payload_script.SCRIPT_SPEND:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraSpend)
				extra = &json_Only_TransactionZetherPayloadExtraSpend{
					payloadExtra.SenderSpendPublicKey.EncodeCompressed(),
					payloadExtra.SenderSpendSignature,
				}
			case transaction_zether_payload_script.SCRIPT_ASSET_CREATE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate)
				extra = &json_Only_TransactionZetherPayloadExtraAssetCreate{
					payloadExtra.Asset,
				}
			case transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease)
				extra = &json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease{
					payloadExtra.AssetId,
					payloadExtra.ReceiverPublicKey,
					payloadExtra.Value,
					payloadExtra.AssetSignature,
					payloadExtra.AssetSupplyPublicKey,
				}
			case transaction_zether_payload_script.SCRIPT_PLAIN_ACCOUNT_FUND:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraPlainAccountFund)
				extra = &json_Only_TransactionZetherPayloadExtraPlainAccountFund{
					payloadExtra.PlainAccountPublicKey,
				}
			case transaction_zether_payload_script.SCRIPT_CONDITIONAL_PAYMENT:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraConditionalPayment)
				extra = &json_Only_TransactionZetherPayloadExtraConditionalPayment{
					payloadExtra.Deadline,
					payloadExtra.DefaultResolution,
					payloadExtra.MultisigThreshold,
					payloadExtra.MultisigPublicKeys,
				}
			default:
				return nil, errors.New("Invalid zether.TxScript")
			}

			payloadsJson[i] = &json_Only_TransactionPayload{
				payload.PayloadScript,
				payload.Asset,
				payload.BurnValue,
				payload.DataVersion,
				payload.Data,
				registrations,
				payload.Parity,
				statementJson,
				payload.WhisperSender,
				payload.WhisperRecipient,
				payload.FeeRate,
				payload.FeeLeadingZeros,
				proofJson,
				extra,
			}

		}

		zetherJson := &json_TransactionZether{
			txJson,
			&json_Only_TransactionZether{
				base.ChainHeight,
				base.ChainKernelHash,
				payloadsJson,
			},
		}

		return marshal(zetherJson)
	default:
		return nil, errors.New("Invalid Tx Version")
	}

}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	return marshalJSON(tx, json.Marshal)
}

func (tx *Transaction) EncodeMsgpack(enc *msgpack.Encoder) error {
	bytes, err := marshalJSON(tx, json.Marshal)
	if err != nil {
		return err
	}

	return enc.EncodeBytes(bytes)
}

func (tx *Transaction) UnmarshalJSON(data []byte) (err error) {

	txOnlyJson := &Json_Transaction{}
	if err = json.Unmarshal(data, txOnlyJson); err != nil {
		return
	}

	switch txOnlyJson.Version {
	case transaction_type.TX_SIMPLE, transaction_type.TX_ZETHER:
	default:
		return errors.New("Invalid Version")
	}

	tx.Version = txOnlyJson.Version
	tx.SpaceExtra = txOnlyJson.SpaceExtra

	switch tx.Version {
	case transaction_type.TX_SIMPLE:

		simpleJson := &json_TransactionSimple{}
		if err = json.Unmarshal(data, simpleJson); err != nil {
			return
		}

		switch simpleJson.DataVersion {
		case transaction_data.TX_DATA_NONE:
			if simpleJson.Data != nil {
				return errors.New("tx.Data must be nil")
			}
		case transaction_data.TX_DATA_PLAIN_TEXT, transaction_data.TX_DATA_ENCRYPTED:
			if simpleJson.Data == nil || len(simpleJson.Data) == 0 || len(simpleJson.Data) > config.TRANSACTIONS_MAX_DATA_LENGTH {
				return errors.New("Invalid tx.Data length")
			}
		default:
			return errors.New("Invalid tx.DataVersion")
		}

		vin := &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: simpleJson.Vin.PublicKey,
			Signature: simpleJson.Vin.Signature,
		}

		base := &transaction_simple.TransactionSimple{
			nil,
			nil,
			simpleJson.TxScript,
			simpleJson.DataVersion,
			simpleJson.Data,
			simpleJson.Nonce,
			simpleJson.Fee,
			vin,
			nil,
		}
		tx.TransactionBaseInterface = base

		switch simpleJson.TxScript {
		case transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
			extraJson := &json_Only_TransactionSimpleExtraUpdateAssetFeeLiquidity{}
			if err = json.Unmarshal(data, extraJson); err != nil {
				return
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity{nil,
				extraJson.Liquidities,
				extraJson.NewCollector,
				extraJson.Collector,
			}
		case transaction_simple.SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT:
			extraJson := &json_Only_TransactionSimpleExtraResolutionConditionalPayment{}
			if err = json.Unmarshal(data, extraJson); err != nil {
				return
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment{nil,
				extraJson.TxId,
				extraJson.PayloadIndex,
				extraJson.Resolution,
				extraJson.MultisigPublicKeys,
				extraJson.Signatures,
			}
		default:
			return errors.New("Invalid json Simple TxScript")
		}

	case transaction_type.TX_ZETHER:

		simpleZether := &json_Only_TransactionZether{}
		if err = json.Unmarshal(data, simpleZether); err != nil {
			return
		}

		payloads := make([]*transaction_zether_payload.TransactionZetherPayload, len(simpleZether.Payloads))
		for i, payload := range simpleZether.Payloads {

			statement := &crypto.Statement{
				RingSize: payload.Statement.RingSize,
				Fee:      payload.Statement.Fee,
			}

			if statement.CLn, err = helpers.ConvertToBN256Array(payload.Statement.CLn); err != nil {
				return
			}
			if statement.CRn, err = helpers.ConvertToBN256Array(payload.Statement.CRn); err != nil {
				return
			}
			if statement.Publickeylist, err = helpers.ConvertToBN256Array(payload.Statement.Publickeylist); err != nil {
				return
			}
			if statement.C, err = helpers.ConvertToBN256Array(payload.Statement.C); err != nil {
				return
			}

			statement.D = new(bn256.G1)
			if err = statement.D.DecodeCompressed(payload.Statement.D); err != nil {
				return err
			}

			m := int(math.Log2(float64(payload.Statement.RingSize)))
			if math.Pow(2, float64(m)) != float64(payload.Statement.RingSize) {
				return errors.New("log failed")
			}

			proof := &crypto.Proof{}
			if err = proof.Deserialize(advanced_buffers.NewBufferReader(payload.Proof), m); err != nil {
				return err
			}

			payloads[i] = &transaction_zether_payload.TransactionZetherPayload{
				payload.PayloadScript,
				payload.Asset,
				payload.BurnValue,
				payload.DataVersion,
				payload.Data,
				&transaction_zether_registrations.TransactionZetherDataRegistrations{
					Registrations: make([]*transaction_zether_registration.TransactionZetherDataRegistration, len(payload.Registrations)),
				},
				payload.Parity,
				statement,
				payload.WhisperSender,
				payload.WhisperRecipient,
				payload.FeeRate,
				payload.FeeLeadingZeros,
				proof,
				nil,
			}

			for i, reg := range payload.Registrations {
				if reg != nil {
					payloads[i].Registrations.Registrations[i] = &transaction_zether_registration.TransactionZetherDataRegistration{
						reg.RegistrationType,
						reg.RegistrationStaked,
						reg.RegistrationSpendPublicKey,
						reg.RegistrationSignature,
					}
				}
			}

			switch payload.PayloadScript {
			case transaction_zether_payload_script.SCRIPT_TRANSFER:
			case transaction_zether_payload_script.SCRIPT_STAKING:
				extraJson := &json_Only_TransactionZetherPayloadExtraStaking{}
				if err = json.Unmarshal(data, extraJson); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraStaking{
					nil,
				}

			case transaction_zether_payload_script.SCRIPT_STAKING_REWARD:
				extraJson := &json_Only_TransactionZetherPayloadExtraStakingReward{}
				if err = json.Unmarshal(data, extraJson); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraStakingReward{
					nil,
					extraJson.Reward,
					extraJson.TemporaryAccountRegistrationIndex,
				}
			case transaction_zether_payload_script.SCRIPT_SPEND:
				extraJson := &json_Only_TransactionZetherPayloadExtraSpend{}
				if err = json.Unmarshal(data, extraJson); err != nil {
					return err
				}

				senderSpendPublicKey := new(bn256.G1)
				if err = senderSpendPublicKey.DecodeCompressed(extraJson.SenderSpendPublicKey); err != nil {
					return
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraSpend{
					nil,
					senderSpendPublicKey,
					extraJson.SenderSpendSignature,
				}
			case transaction_zether_payload_script.SCRIPT_ASSET_CREATE:
				extraJson := &json_Only_TransactionZetherPayloadExtraAssetCreate{}
				if err = json.Unmarshal(data, extraJson); err != nil {
					return err
				}
				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate{
					Asset: extraJson.Asset,
				}
			case transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE:
				extraJson := &json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease{}
				if err = json.Unmarshal(data, extraJson); err != nil {
					return err
				}
				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease{
					nil,
					extraJson.AssetId,
					extraJson.ReceiverPublicKey,
					extraJson.Value,
					extraJson.AssetSignature,
					extraJson.AssetSupplyPublicKey,
				}
			case transaction_zether_payload_script.SCRIPT_PLAIN_ACCOUNT_FUND:
				extraJson := &json_Only_TransactionZetherPayloadExtraPlainAccountFund{}
				if err = json.Unmarshal(data, extraJson); err != nil {
					return err
				}
				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraPlainAccountFund{
					nil,
					extraJson.PlainAccountPublicKey,
				}
			case transaction_zether_payload_script.SCRIPT_CONDITIONAL_PAYMENT:
				extraJson := &json_Only_TransactionZetherPayloadExtraConditionalPayment{}
				if err = json.Unmarshal(data, extraJson); err != nil {
					return err
				}
				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraConditionalPayment{
					nil,
					extraJson.Deadline,
					extraJson.DefaultResolution,
					extraJson.MultisigThreshold,
					extraJson.MultisigPublicKeys,
				}
			default:
				return errors.New("Invalid Zether TxScript")
			}

		}

		base := &transaction_zether.TransactionZether{
			ChainHeight:     simpleZether.ChainHeight,
			ChainKernelHash: simpleZether.ChainKernelHash,
			Payloads:        payloads,
		}

		tx.TransactionBaseInterface = base

	default:
		return errors.New("Invalid Version")
	}

	return nil
}

func (tx *Transaction) DecodeMsgpack(dec *msgpack.Decoder) error {
	bytes, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	return tx.UnmarshalJSON(bytes)
}
