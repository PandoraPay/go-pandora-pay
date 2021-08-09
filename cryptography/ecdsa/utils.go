package ecdsa

// DigestLength sets the signature digest exact length
const DigestLength = 32

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}
