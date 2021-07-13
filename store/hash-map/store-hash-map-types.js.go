package hash_map

import "pandora-pay/helpers"

type ChangesMapElement struct {
	Element helpers.SerializableInterface
	Status  string
}

type CommittedMapElement struct {
	Element helpers.SerializableInterface
	Status  string
	Stored  string
}
