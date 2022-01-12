package config

type SeedNode struct {
	Url string `json:"url" msgpack:"url"`
}

var (
	MAIN_NET_SEED_NODES = []*SeedNode{}

	TEST_NET_SEED_NODES = []*SeedNode{
		{
			"ws://helloworldx.ddns.net:16000/ws",
		},
		{
			"ws://helloworldx.ddns.net:16001/ws",
		},
		{
			"ws://helloworldx.ddns.net:16002/ws",
		},
		{
			"ws://helloworldx.ddns.net:16003/ws",
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
