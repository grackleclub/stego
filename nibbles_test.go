package cryptogif

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytesAndNibbles(t *testing.T) {	
	for secretLen := range 1028 {
		t.Run(fmt.Sprintf("bytesNibs-%d", secretLen+1), func(t *testing.T) {
			t.Parallel()
			secret, err := newSecret(secretLen + 1)
			require.NoError(t, err)

			nibbles, err := toNibbles(secret)
			require.NoError(t, err)

			bytes, err := toBytes(nibbles)
			require.NoError(t, err)

			require.Equal(t, len(secret)*2, len(nibbles))
			require.Equal(t, secret, bytes)
			t.Logf("Expect: %x\nActual: %x\n", secret, bytes)
		})

	}
}

func TestSpecialCases(t *testing.T) {
	t.Run("odd characters", func(t *testing.T) {
		t.Parallel()
		original := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x3c}
		nibbles, err := toNibbles(original)
		require.NoError(t, err)
		require.NotNil(t, nibbles)

		bytes, err := toBytes(nibbles)
		require.NoError(t, err)
		require.NotNil(t, bytes)

		require.Equal(t, string(original), string(bytes))
		t.Logf("\nExpect: %s\nActual: %s\n", original, bytes)
	})
	t.Run("special characters", func(t *testing.T) {
		t.Parallel()
		original := []byte("hello <> world")
		nibbles, err := toNibbles(original)
		require.NoError(t, err)
		require.NotNil(t, nibbles)

		bytes, err := toBytes(nibbles)
		require.NoError(t, err)
		require.NotNil(t, bytes)

		require.Equal(t, string(original), string(bytes))
		t.Logf("\nExpect: %s\nActual: %s\n", original, bytes)
	})
}
