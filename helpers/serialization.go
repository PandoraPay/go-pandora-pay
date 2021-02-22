package helpers

import "errors"

// serialize number
func SerializeNumber(n uint64) (b []byte) {

	for n >= 0x80 {

		c := n & 0x7f
		b = append(b, byte(c|0x80))

		n = (n - c) >> 7
	}
	b = append(b, byte(n))

	return
}

// Variable-Length Encoding of Integers based on ReadVarInt
func UnserializeNumber(b []byte, pos *int) (r uint64, err error) {

	power := uint64(1)
	r = 0

	for *pos < len(b) {

		a := b[*pos]
		*pos += 1

		r += uint64(a&0x7f) * power
		if a&0x80 == 0 {
			return
		}
		power <<= 7
	}

	return
}

func UnserializeBuffer(b []byte, pos *int, count int) (result []byte, err error) {
	if *pos+count >= len(b) {
		err = errors.New("Buffer exceeded")
		return
	}

	copy(result[:], b[*pos:*pos+count])
	*pos += count
	return
}
