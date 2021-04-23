package api_common

type APIBlockchain struct {
	Height          uint64
	Hash            string
	PrevHash        string
	KernelHash      string
	PrevKernelHash  string
	Timestamp       uint64
	Transactions    uint64
	Target          string
	TotalDifficulty string
}
