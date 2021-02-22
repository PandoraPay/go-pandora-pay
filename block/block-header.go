package block

type BlockHeader struct {
	Version   uint64
	Timestamp uint64
	Height    uint64
}

func (blockHeader *BlockHeader) Serialize() ([]byte, error) {

}

func (blockHeader *BlockHeader) Deserialize(b []byte) error {

}
