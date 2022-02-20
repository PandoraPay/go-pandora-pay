package transaction

import (
	"encoding/json"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
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
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations/transaction_zether_registration"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type json_TransactionDataRegistration struct {
	RegistrationType      transaction_zether_registration.TransactionZetherDataRegistrationType `json:"registrationType"  msgpack:"registrationType"`
	RegistrationSignature []byte
}

type json_TransactionDataDelegatedStakingUpdate struct {
	DelegatedStakingHasNewInfo   bool   `json:"delegatedStakingHasNewInfo" msgpack:"delegatedStakingHasNewInfo"`
	DelegatedStakingNewPublicKey []byte `json:"delegatedStakingNewPublicKey" msgpack:"delegatedStakingNewPublicKey"` //20 byte
	DelegatedStakingNewFee       uint64 `json:"delegatedStakingNewFee" msgpack:"delegatedStakingNewFee"`
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
	FeeVersion  bool                                    `json:"feeVersion" msgpack:"feeVersion"`
	Vin         *json_TransactionSimpleInput            `json:"vin" msgpack:"vin"`
	Extra       interface{}                             `json:"extra" msgpack:"extra"`
}

type json_TransactionSimpleInput struct {
	PublicKey []byte `json:"publicKey,omitempty" msgpack:"publicKey,omitempty"` //32
	Signature []byte `json:"signature" msgpack:"signature"`                     //64
}

type json_Only_TransactionSimpleExtraUpdateDelegate struct {
	DelegatedStakingClaimAmount uint64                                      `json:"delegatedStakingClaimAmount"  msgpack:"delegatedStakingClaimAmount"`
	DelegatedStakingUpdate      *json_TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"  msgpack:"delegatedStakingUpdate"`
}

type json_Only_TransactionSimpleExtraUnstake struct {
	Amount uint64 `json:"amount"  msgpack:"amount"`
}

type json_Only_TransactionSimpleExtraUpdateAssetFeeLiquidity struct {
	Liquidities     []*asset_fee_liquidity.AssetFeeLiquidity `json:"liquidities"`
	CollectorHasNew bool                                     `json:"collectorHasNew"`
	Collector       []byte                                   `json:"collector"`
}

type json_Only_TransactionZether struct {
	ChainHeight uint64                          `json:"chainHeight"  msgpack:"chainHeight"`
	ChainHash   []byte                          `json:"chainHash"  msgpack:"chainHash"`
	Payloads    []*json_Only_TransactionPayload `json:"payloads"  msgpack:"payloads"`
}

type json_Only_TransactionZetherPayloadExtraDelegateStake struct {
	DelegatePublicKey      []byte                                      `json:"delegatePublicKey"  msgpack:"delegatePublicKey"`
	ConvertToUnclaimed     bool                                        `json:"convertToUnclaimed"  msgpack:"convertToUnclaimed"`
	DelegatedStakingUpdate *json_TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"  msgpack:"delegatedStakingUpdate"`
	DelegateSignature      []byte                                      `json:"delegateSignature"  msgpack:"delegateSignature"`
}

type json_Only_TransactionZetherPayloadExtraStakingReward struct {
	Reward                            uint64 `json:"reward"  msgpack:"reward"`
	TemporaryAccountRegistrationIndex uint64 `json:"temporaryAccountRegistrationIndex"  msgpack:"temporaryAccountRegistrationIndex"`
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
	PayloadScript    transaction_zether_payload.PayloadScriptType `json:"payloadScript"  msgpack:"payloadScript"`
	Asset            []byte                                       `json:"asset"  msgpack:"asset"`
	BurnValue        uint64                                       `json:"burnValue"  msgpack:"burnValue"`
	DataVersion      transaction_data.TransactionDataVersion      `json:"dataVersion"  msgpack:"dataVersion"`
	Data             []byte                                       `json:"data"  msgpack:"data"`
	Registrations    []*json_TransactionDataRegistration          `json:"registrations"  msgpack:"registrations"`
	Parity           bool                                         `json:"parity" msgpack:"parity"`
	Statement        *json_Only_TransactionZetherStatement        `json:"statement"  msgpack:"statement"`
	WhisperSender    []byte                                       `json:"whisperSender" msgpack:"whisperSender"`
	WhisperRecipient []byte                                       `json:"whisperRecipient" msgpack:"whisperRecipient"`
	FeeRate          uint64                                       `json:"feeRate"  msgpack:"feeRate"`
	FeeLeadingZeros  byte                                         `json:"feeLeadingZeros"  msgpack:"feeLeadingZeros"`
	Proof            []byte                                       `json:"proof"  msgpack:"proof"`
	Extra            interface{}                                  `json:"extra"  msgpack:"extra"`
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

		vinJson := &json_TransactionSimpleInput{
			base.Vin.PublicKey,
			base.Vin.Signature,
		}

		simpleJson := &json_TransactionSimple{
			txJson,
			base.TxScript,
			base.DataVersion,
			base.Data,
			base.Nonce,
			base.Fee,
			base.FeeVersion,
			vinJson,
			nil,
		}

		switch base.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUpdateDelegate)
			simpleJson.Extra = &json_Only_TransactionSimpleExtraUpdateDelegate{
				extra.DelegatedStakingClaimAmount,
				&json_TransactionDataDelegatedStakingUpdate{
					extra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
					extra.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
					extra.DelegatedStakingUpdate.DelegatedStakingNewFee,
				},
			}
		case transaction_simple.SCRIPT_UNSTAKE:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUnstake)
			simpleJson.Extra = json_Only_TransactionSimpleExtraUnstake{
				extra.Amount,
			}
		case transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity)
			simpleJson.Extra = json_Only_TransactionSimpleExtraUpdateAssetFeeLiquidity{
				extra.Liquidities,
				extra.CollectorHasNew,
				extra.Collector,
			}
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

			w := helpers.NewBufferWriter()
			payload.Proof.Serialize(w)
			proofJson := w.Bytes()

			var extra interface{}

			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_TRANSFER:
				//no payload
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake)
				extra = &json_Only_TransactionZetherPayloadExtraDelegateStake{
					payloadExtra.DelegatePublicKey,
					payloadExtra.ConvertToUnclaimed,
					&json_TransactionDataDelegatedStakingUpdate{
						payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
						payloadExtra.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
						payloadExtra.DelegatedStakingUpdate.DelegatedStakingNewFee,
					},
					payloadExtra.DelegateSignature,
				}
			case transaction_zether_payload.SCRIPT_STAKING_REWARD:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraStakingReward)
				extra = &json_Only_TransactionZetherPayloadExtraStakingReward{
					payloadExtra.Reward,
					payloadExtra.TemporaryAccountRegistrationIndex,
				}
			case transaction_zether_payload.SCRIPT_ASSET_CREATE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate)
				extra = &json_Only_TransactionZetherPayloadExtraAssetCreate{
					payloadExtra.Asset,
				}
			case transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease)
				extra = &json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease{
					payloadExtra.AssetId,
					payloadExtra.ReceiverPublicKey,
					payloadExtra.Value,
					payloadExtra.AssetSignature,
					payloadExtra.AssetSupplyPublicKey,
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
				base.ChainHash,
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
			simpleJson.FeeVersion,
			vin,
			nil,
		}
		tx.TransactionBaseInterface = base

		switch simpleJson.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE:

			extraJson := &json_Only_TransactionSimpleExtraUpdateDelegate{}
			if err = json.Unmarshal(data, extraJson); err != nil {
				return
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateDelegate{
				nil,
				extraJson.DelegatedStakingClaimAmount,
				&transaction_data.TransactionDataDelegatedStakingUpdate{
					extraJson.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
					extraJson.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
					extraJson.DelegatedStakingUpdate.DelegatedStakingNewFee,
				},
			}

		case transaction_simple.SCRIPT_UNSTAKE:
			extraJson := &json_Only_TransactionSimpleExtraUnstake{}
			if err = json.Unmarshal(data, extraJson); err != nil {
				return
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleExtraUnstake{
				Amount: extraJson.Amount,
			}
		case transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
			extraJson := &json_Only_TransactionSimpleExtraUpdateAssetFeeLiquidity{}
			if err = json.Unmarshal(data, extraJson); err != nil {
				return
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity{
				nil,
				extraJson.Liquidities,
				extraJson.CollectorHasNew,
				extraJson.Collector,
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
			if err = proof.Deserialize(helpers.NewBufferReader(payload.Proof), m); err != nil {
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
						reg.RegistrationSignature,
					}
				}
			}

			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_TRANSFER:
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				extraJson := &json_Only_TransactionZetherPayloadExtraDelegateStake{}
				if err := json.Unmarshal(data, extraJson); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake{
					nil,
					extraJson.DelegatePublicKey,
					extraJson.ConvertToUnclaimed,
					&transaction_data.TransactionDataDelegatedStakingUpdate{
						extraJson.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
						extraJson.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
						extraJson.DelegatedStakingUpdate.DelegatedStakingNewFee,
					},
					extraJson.DelegateSignature,
				}

			case transaction_zether_payload.SCRIPT_STAKING_REWARD:
				extraJson := &json_Only_TransactionZetherPayloadExtraStakingReward{}
				if err := json.Unmarshal(data, extraJson); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraStakingReward{
					nil,
					extraJson.Reward,
					extraJson.TemporaryAccountRegistrationIndex,
				}

			case transaction_zether_payload.SCRIPT_ASSET_CREATE:
				extraJson := &json_Only_TransactionZetherPayloadExtraAssetCreate{}
				if err := json.Unmarshal(data, extraJson); err != nil {
					return err
				}
				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate{
					Asset: extraJson.Asset,
				}
			case transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
				extraJson := &json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease{}
				if err := json.Unmarshal(data, extraJson); err != nil {
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

			default:
				return errors.New("Invalid Zether TxScript")
			}

		}

		base := &transaction_zether.TransactionZether{
			ChainHeight: simpleZether.ChainHeight,
			ChainHash:   simpleZether.ChainHash,
			Payloads:    payloads,
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
