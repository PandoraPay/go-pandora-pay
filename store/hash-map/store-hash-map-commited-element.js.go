package hash_map

import "pandora-pay/helpers"

type CommittedMapElement struct {
	Element helpers.SerializableInterface
	Status  string
	Stored  string
}
