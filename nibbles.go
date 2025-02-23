package cryptogif

import "log/slog"

func splitNibbles(b byte) (byte, byte) {
	return b >> 4, b & 0x0f
}

func joinNibbles(n1, n2 byte) byte {
	return n1<<4 | n2
}

// toNibbles takes a byte slice and returns a slice of nibbles,
// guaranteed to be even in length.
func toNibbles(bytes []byte) []uint8 {
	var nibbles []uint8
	for _, b := range bytes {
		n1, n2 := splitNibbles(b)
		nibbles = append(nibbles, n1, n2)
	}
	slog.Debug("converted bytes to nibbles", "len_nib", len(nibbles), "len_bytes", len(bytes))
	return nibbles
}

// toBytes takes an even length slice of nibbles and returns a byte slice.
func toBytes(nibbles []uint8) []byte {
	var stretched []byte

	// TODO fix this hack
	// if len(nibbles)%2 != 0 {
	// 	slog.Warn("odd number of nibbles, dropping last", "len", len(nibbles))
	// 	nibbles = nibbles[:len(nibbles)-1]
	// }

	for i := 0; i < len(nibbles); i += 2 {
		if i+1 < len(nibbles) {
			stretched = append(stretched, joinNibbles(nibbles[i], nibbles[i+1]))
		}
		// if len(nibbles) <= i {
		// 	stretched = append(stretched, nibbles[i])
		// 	return stretched
		// } else {
		// 	stretched = append(stretched, joinNibbles(nibbles[i], nibbles[i+1]))
		// }
	}
	return stretched
}
