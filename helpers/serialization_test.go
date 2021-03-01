package helpers

import (
	"encoding/binary"
	"testing"
)

func TestSerializeNumber(t *testing.T) {

	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, 0)
	if n != 1 || b[0] != 0 {
		t.Errorf("Invalid serialization %d %d %s", len(b), b[0], string(b))
	}

	n = binary.PutUvarint(b, 1)
	if n != 1 || b[0] != 1 {
		t.Errorf("Invalid serialization %d %d %s", len(b), b[0], string(b))
	}

	n = binary.PutUvarint(b, 120)
	if n != 1 || b[0] != 120 {
		t.Errorf("Invalid serialization %d %d %s", len(b), b[0], string(b))
	}

	n = binary.PutUvarint(b, 126)
	if n != 1 || b[0] != 126 {
		t.Errorf("Invalid serialization %d %d %s", len(b), b[0], string(b))
	}

	n = binary.PutUvarint(b, 127)
	if n != 1 || b[0] != 127 {
		t.Errorf("Invalid serialization %d %d %s", len(b), b[0], string(b))
	}

	n = binary.PutUvarint(b, 128)
	if n != 2 || b[0] != 128 {
		t.Errorf("Invalid serialization %d %d %s", len(b), b[0], string(b))
	}

	a, done := binary.Uvarint(b)
	if a != 128 || done <= 0 {
		t.Errorf("Invalid serialization %d %d", a, b)
	}

	binary.PutUvarint(b, 0xFFFFFFFFFFFFFFFF)
	a, done = binary.Uvarint(b)
	if a != 0xFFFFFFFFFFFFFFFF || done <= 0 {
		t.Errorf("Invalid serialization %d %d", b, a)
	}

	binary.PutUvarint(b, 0xFFFFFFFFFFFFFFFC)
	a, done = binary.Uvarint(b)
	if a != 0xFFFFFFFFFFFFFFFC {
		t.Errorf("Invalid serialization %d %d", a, b)
	}

}

func TestDeserializeNumber(t *testing.T) {

	b := make([]byte, binary.MaxVarintLen64)
	for i := 0; i < 100; i++ {

		no := RandomUint64()

		binary.PutUvarint(b, no)

		a, done := binary.Uvarint(b)
		if a != no || done <= 0 {
			t.Errorf("Invalid serialization deserialization %d %s %d", no, string(b), a)
		}
	}

}
