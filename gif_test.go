package stego

import (
	"crypto/rand"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testSource = path.Join("img", "originals", "earth.gif")
	testWrite  = path.Join("img", "output", "earth_write.gif")
	testOutput = path.Join("img", "output", "earth_output.gif")
)

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
	bytes, err := newSecret(64) // TODO fails with 128
	require.NoError(t, err)

	t.Run("validate", func(t *testing.T) {
		actual, err := encodeDecode(testSource, bytes)
		require.NoError(t, err)
		require.Equal(t, string(bytes), string(actual))
	})
}

func TestLifecycle(t *testing.T) {
	g, err := Read(testSource)
	require.NoError(t, err)

	t.Run("preview", func(t *testing.T) {
		text := "Hello, world!"
		g, err := Inject(g, []byte(text))
		require.NoError(t, err)
		require.NotNil(t, g)

		err = Write(g, testOutput)
		require.NoError(t, err)

		new, err := Read(testOutput)
		require.NoError(t, err)
		require.NotNil(t, new)

		data, err := Extract(new)
		t.Logf("output (str): %s\n", data)
		require.NoError(t, err)
		require.Equal(t, text, string(data))
	})
}

// encodeDecode reads a gif, encodes a random secret, then decodes it,
// returning the input and output secrets for comparison in a test context.
func encodeDecode(path string, input []byte) (string, error) {
	g, err := Read(path)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}
	gNew, err := Inject(g, input)
	if err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}
	output, err := Extract(gNew)
	if err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return string(output), nil
}
