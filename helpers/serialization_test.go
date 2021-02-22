package helpers

import (
	"testing"
)

func TestSerializeNumber(t *testing.T) {

	a := SerializeNumber(0)
	if len(a) != 1 || a[0] != 0 {
		t.Errorf("Invalid serialization %d %d %s", len(a), a[0], string(a))
	}

	a = SerializeNumber(1)
	if len(a) != 1 || a[0] != 1 {
		t.Errorf("Invalid serialization %d %d %s", len(a), a[0], string(a))
	}

	a = SerializeNumber(2)
	if len(a) != 1 || a[0] != 2 {
		t.Errorf("Invalid serialization %d %d %s", len(a), a[0], string(a))
	}

	a = SerializeNumber(120)
	if len(a) != 1 || a[0] != 120 {
		t.Errorf("Invalid serialization %d %d %s", len(a), a[0], string(a))
	}

	a = SerializeNumber(126)
	if len(a) != 1 || a[0] != 126 {
		t.Errorf("Invalid serialization %d %d %s", len(a), a[0], string(a))
	}

	a = SerializeNumber(127)
	if len(a) != 1 || a[0] != 127 {
		t.Errorf("Invalid serialization %d %d %s", len(a), a[0], string(a))
	}

	a = SerializeNumber(128)
	if len(a) != 2 || a[0] != 128 {
		t.Errorf("Invalid serialization %d %d %s", len(a), a[0], string(a))
	}

	pos := 0
	b, err := UnserializeNumber(a, &pos)
	if b != 128 || err != nil {
		t.Errorf("Invalid serialization %d %d", a, b)
	}

	a = SerializeNumber(0xFFFFFFFFFFFFFFFF)
	pos = 0
	b, err = UnserializeNumber(a, &pos)
	if b != 0xFFFFFFFFFFFFFFFF || err != nil {
		t.Errorf("Invalid serialization %d %d", a, b)
	}

	a = SerializeNumber(0xFFFFFFFFFFFFFFFC)
	pos = 0
	b, err = UnserializeNumber(a, &pos)
	if b != 0xFFFFFFFFFFFFFFFC || err != nil {
		t.Errorf("Invalid serialization %d %d", a, b)
	}

}

func TestUnserializeNumber(t *testing.T) {

	for i := 0; i < 100; i++ {

		no := Uint64()

		a := SerializeNumber(no)

		pos := 0
		b, err := UnserializeNumber(a, &pos)
		if b != no || err != nil {
			t.Errorf("Invalid serialization deserialization %d %d %s", no, b, string(a))
		}
	}

}
