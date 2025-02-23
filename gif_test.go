package cryptogif

import (
	"encoding/base64"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testSource = path.Join("img", "earth.gif")
	testDest   = path.Join("img", "earth_out.gif")
)

func TestRead(t *testing.T) {
	_, err := Read(path.Join("img", "wiki.gif"))
	require.NoError(t, err)
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
	secret := "hello-there "

	b64 := base64.StdEncoding.EncodeToString([]byte(secret))

	err := encode([]byte(b64), testSource, testDest)
	require.NoError(t, err)

	resultBytes, err := decode(testDest)
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
