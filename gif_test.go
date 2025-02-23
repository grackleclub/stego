package cryptogif

import (
	"crypto/rand"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testSource = path.Join("img", "earth.gif")
	testDest   = path.Join("img", "earth_out.gif")
)

func TestRead(t *testing.T) {
	_, err := Read(testSource)
	require.NoError(t, err)
}

// newSecret generates a random secret of the given length, for testing.
func newSecret(length int) ([]byte, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	return secret, nil
}

func TestNewSecret(t *testing.T) {
	secret, err := newSecret(32)
	require.NoError(t, err)
	require.Len(t, secret, 32)
	t.Logf("new secret: %x\n", secret)
}

func TestEncode(t *testing.T) {

	tests := 3
	len := 8
	for i := range tests {
		l := len * (i + 1)
		secret, err := newSecret(l)
		require.NoError(t, err)

		g, err := Read(testSource)
		require.NoError(t, err)

		gNew, err := Encode(g, secret)
		require.NoError(t, err)

		text, err := Decode(gNew)
		require.NoError(t, err)
		require.Equal(t, secret, text)
	}
}

func TestNewPI(t *testing.T) {
	g, err := Read(path.Join("img", "earth.gif"))
	require.NoError(t, err)
	_, err = newPaletteInfo(g)
	require.NoError(t, err)
}
