package helpers

func SerializeBoolToByte(value bool) byte {
	if value {
		return 1
	}
	return 0
}
