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
	tests := 10

	for testNum := range tests {
		t.Run(fmt.Sprintf("test-%d", testNum), func(t *testing.T) {
			t.Parallel()
			expect, actual, err := EncodeDecode(testSource, 0+testNum)
			require.NoError(t, err)
			require.Equal(t, expect, actual)
		})
	}
}

func TestNewPI(t *testing.T) {
	g, err := Read(path.Join("img", "earth.gif"))
	require.NoError(t, err)
	_, err = newPaletteInfo(g)
	require.NoError(t, err)
}

// EncodeDecode reads a gif, encodes a random secret, then decodes it,
// returning the input and output secrets for comparison in a test context.
func EncodeDecode(path string, len int) ([]byte, []byte, error) {
	g, err := Read(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read: %w", err)
	}
	input, err := newSecret(len)
	if err != nil {
		return nil, nil, fmt.Errorf("new secret: %w", err)
	}
	gNew, err := Encode(g, input)
	if err != nil {
		return nil, nil, fmt.Errorf("encode: %w", err)
	}
	output, err := Decode(gNew)
	if err != nil {
		return nil, nil, fmt.Errorf("decode: %w", err)
	}
	return input, output, nil
}
