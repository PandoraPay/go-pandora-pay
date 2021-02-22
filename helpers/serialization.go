package helpers

// serialize number
func SerializeNumber(n uint64) []byte {

	var b []byte

	for n >= 0x80 {

		c := n & 0x7f
		b = append(b, byte(c|0x80))

		n = (n - c) >> 7
	}
	b = append(b, byte(n))

	return b
}

// Variable-Length Encoding of Integers based on ReadVarInt
func UnserializeNumber(b []byte, pos *int) uint64 {

	power := uint64(1)
	r := uint64(0)

	for *pos < len(b) {

		a := b[*pos]
		*pos += 1

		r += uint64(a&0x7f) * power
		if a&0x80 == 0 {
			return r
		}
		power <<= 7
	}

	return r
}
