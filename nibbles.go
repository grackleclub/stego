package cryptogif

import (
	"fmt"
	"log/slog"
)

func splitNibbles(b byte) (byte, byte) {
	return b >> 4, b & 0x0f
}

func joinNibbles(n1, n2 byte) byte {
	return n1<<4 | n2
}

// toNibbles takes a byte slice and returns a slice of nibbles,
// guaranteed to be even in length.
func toNibbles(bytes []byte) ([]uint8, error) {
	if len(bytes) == 0 {
		return nil, fmt.Errorf("empty byte slice")
	}
	var nibbles []uint8
	for _, b := range bytes {
		n1, n2 := splitNibbles(b)
		nibbles = append(nibbles, n1, n2)
	}
	slog.Debug("converted bytes to nibbles", "len_nib", len(nibbles), "len_bytes", len(bytes), "ratio", len(nibbles)/len(bytes))
	return nibbles, nil
}

// toBytes takes an even length slice of nibbles and returns a byte slice.
func toBytes(nibbles []uint8) ([]byte, error) {
	if len(nibbles) == 0 {
		return nil, fmt.Errorf("empty nibble slice")
	}
	if len(nibbles)%2 != 0 {
		return nil, fmt.Errorf("odd length nibble slice")
	}

	var bytes []byte
	for i := 0; i < len(nibbles); i += 2 {
		if i+1 < len(nibbles) {
			bytes = append(bytes, joinNibbles(nibbles[i], nibbles[i+1]))
		}
	}
	return bytes, nil
}
