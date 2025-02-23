package cryptogif

import (
	"crypto/rand"
	"encoding/base64"
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
	// secret, err := newSecret(16)
	// require.NoError(t, err)
	secret := "hello  there"

	b64 := base64.StdEncoding.EncodeToString([]byte(secret))

	err := Encode([]byte(b64), testSource, testDest)
	require.NoError(t, err)

	resultBytes, err := Decode(testDest)
	require.NoError(t, err)

	var result = make([]byte, len(resultBytes))
	n, err := base64.StdEncoding.Decode(result, resultBytes)
	require.NoError(t, err)
	result = result[:n]

	t.Logf("original: %v", b64)
	t.Logf("result: %v", string(resultBytes))
	t.Logf("decoded: %v", string(result))

	require.Equal(t, b64, string(resultBytes))
}

func TestNewPI(t *testing.T) {
	g, err := Read(path.Join("img", "earth.gif"))
	require.NoError(t, err)
	_, err = newPaletteInfo(g)
	require.NoError(t, err)
}
