package txs_builder_zether_helper

type TxsBuilderZetherTxPayloadBase struct {
	Sender         string `json:"sender" msgpack:"sender"`
	Recipient      string `json:"recipient" msgpack:"recipient"`
	RingSize       int    `json:"ringSize"  msgpack:"ringSize"`
	WitnessIndexes []int  `json:"witnessIndexes" msgpack:"witnessIndexes"`
}

type TxsBuilderZetherTxDataBase struct {
	Payloads []*TxsBuilderZetherTxPayloadBase `json:"payloads" msgpack:"payloads"`
}
