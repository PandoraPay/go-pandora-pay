package subscription

type Subscription struct {
	Id     uint64
	Name   []byte
	Key    []byte
	Option interface{}
}
