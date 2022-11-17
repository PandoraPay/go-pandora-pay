package api_code_types

type APIReturnType uint8

const (
	RETURN_SERIALIZED APIReturnType = iota
	RETURN_JSON
)
