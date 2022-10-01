package wizard

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
)

type WizardTxSimpleExtra interface {
}

type WizardTxSimpleExtraUpdateAssetFeeLiquidity struct {
	WizardTxSimpleExtra `json:"-"  msgpack:"-"`
	Liquidities         []*asset_fee_liquidity.AssetFeeLiquidity `json:"liquidities"  msgpack:"liquidities"`
	NewCollector        bool                                     `json:"newCollector"  msgpack:"newCollector"`
	Collector           []byte                                   `json:"collector"  msgpack:"collector"`
}

type WizardTxSimpleExtraResolutionPayInFuture struct {
	WizardTxSimpleExtra `json:"-"  msgpack:"-"`
	TxId                []byte   `json:"txId" msgpack:"txId"`
	PayloadIndex        byte     `json:"payloadIndex" msgpack:"payloadIndex"`
	Resolution          bool     `json:"resolution" msgpack:"resolution"`
	MultisigPublicKeys  [][]byte `json:"multisigPublicKeys" msgpack:"multisigPublicKeys"`
	Signatures          [][]byte `json:"signatures" msgpack:"signatures"`
}

type WizardTxSimpleTransfer struct {
	Extra WizardTxSimpleExtra    `json:"extra" msgpack:"extra"`
	Data  *WizardTransactionData `json:"data" msgpack:"data"`
	Fee   *WizardTransactionFee  `json:"fee" msgpack:"fee"`
	Nonce uint64                 `json:"nonce" msgpack:"nonce"`
	Key   []byte                 `json:"key" msgpack:"key"`
}
