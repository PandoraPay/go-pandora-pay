package urldecoder

import (
	"github.com/gorilla/schema"
)

var (
	Decoder = schema.NewDecoder()
	Encoder = schema.NewEncoder()
)
