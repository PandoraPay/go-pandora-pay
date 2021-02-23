package config

const (
	NAME    string = "PANDORA PAY"
	VERSION string = "0.0"

	NETWORK_SELECTED uint64 = 0

	MAIN_NET_NETWORK_BYTE        uint64 = 0
	MAIN_NET_NETWORK_BYTE_PREFIX string = "PANDORA" // must have 7 characters

	TEST_NET_NETWORK_BYTE        uint64 = 1033
	TEST_NET_NETWORK_BYTE_PREFIX string = "PANTEST" // must have 7 characters

	DEV_NET_NETWORK_BYTE        uint64 = 4255
	DEV_NET_NETWORK_BYTE_PREFIX string = "PANDDEV" // must have 7 characters

	NETWORK_BYTE_PREFIX_LENGTH = 7

	BLOCK_MAX_SIZE uint64 = 1 << 10
)
