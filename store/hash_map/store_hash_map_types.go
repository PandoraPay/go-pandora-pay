package hash_map

import "pandora-pay/helpers"

type ChangesMapElement struct {
	Element      helpers.SerializableInterface
	Status       string
	index        uint64
	indexProcess bool
}

type CommittedMapElement struct {
	Element    helpers.SerializableInterface
	Status     string
	Stored     string
	serialized []byte
	size       int
}
