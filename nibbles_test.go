package cryptogif

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitAndJoinNibbles(t *testing.T) {
	max := 255
	for i := 0; i < max; i++ {
		n1, n2 := splitNibbles(byte(i))
		b := joinNibbles(n1, n2)
		require.Equal(t, byte(i), b)
		// t.Logf("byte: %d, nibbles: %d, %d\n", i, n1, n2)
	}
}

func TestCrushAndStretch(t *testing.T) {
	secret, err := newSecret(32)
	require.NoError(t, err)
	crushed := toNibbles(secret)
	stretched := toBytes(crushed)
	require.Equal(t, secret, stretched)
	// t.Log("original: ", secret)
	// t.Log("crushed: ", crushed)
	// t.Log("stretched: ", stretched)
	require.Equal(t, len(secret)*2, len(crushed))
}
