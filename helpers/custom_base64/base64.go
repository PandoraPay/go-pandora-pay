package custom_base64

import "encoding/base64"

var Base64Encoder *base64.Encoding

func init() {
	const encodeStd = "ABCDEFGHIJKLMN#PQRSTUVWXYZabcdefghijk@mnopqrstuvwxyz0123456789+$"
	Base64Encoder = base64.NewEncoding(encodeStd)
}
