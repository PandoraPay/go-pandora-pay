package advanced_connection_types

type AdvancedConnectionMessage struct {
	ReplyId     uint32
	ReplyStatus bool
	ReplyAwait  bool
	Name        []byte
	Data        []byte
}

type AdvancedConnectionReply struct {
	Out     []byte
	Err     error
	Timeout bool
}

type UUID uint64

const (
	UUID_ALL      = UUID(0)
	UUID_SKIP_ALL = UUID(1)
)
