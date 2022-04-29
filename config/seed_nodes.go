package config

type SeedNode struct {
	Url string `json:"url" msgpack:"url"`
}

var (
	MAIN_NET_SEED_NODES = []*SeedNode{}

	TEST_NET_SEED_NODES = []*SeedNode{
		{
			"wss://webdexplorer.ddns.net:17000/ws",
		},
		{
			"wss://webdexplorer.ddns.net:17001/ws",
		},
		{
			"wss://webdexplorer.ddns.net:17002/ws",
		},
		{
			"wss://webdexplorer.ddns.net:17003/ws",
		},
	}

	DEV_NET_SEED_NODES = []*SeedNode{
		{
			"ws://127.0.0.1:5230/ws",
		},
		{
			"ws://127.0.0.1:5231/ws",
		},
		{
			"ws://127.0.0.1:5232/ws",
		},
		{
			"ws://127.0.0.1:5233/ws",
		},
	}
)
