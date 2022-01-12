package txs_builder

type ZetherRingConfiguration struct {
	RingSize    int `json:"ringSize"  msgpack:"ringSize"`
	NewAccounts int `json:"newAccounts"  msgpack:"newAccounts"`
}
