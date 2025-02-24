package cryptogif

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testSource  = path.Join("img", "earth.gif")
	testWrite   = path.Join("img", "earth_write.gif")
	testFailLog = path.Join("test.log")
)

func TestRead(t *testing.T) {
	_, err := Read(testSource)
	require.NoError(t, err)
}

func TestWrite(t *testing.T) {
	g, err := Read(testSource)
	require.NoError(t, err)
	err = Write(g, testWrite)
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

// This function has intermittent failures, so I'm isolating if the palette.Index is non-deterministic,
// or if there are certain bytes that can cause errors in encoding/decoding.
func TestRetry(t *testing.T) {
	var priorFailures []string
	f, err := os.Open(testFailLog)
	require.NoError(t, err)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		priorFailures = append(priorFailures, parts[0])
	}

	for i, fail := range priorFailures {
		t.Run(fmt.Sprintf("prior-%d", i), func(t *testing.T) {
			// t.Parallel()
			g, err := Read(testSource)
			require.NoError(t, err)
			g, err = Encode(g, fail)
			require.NoError(t, err)
			bytes, err := Decode(g)
			require.NoError(t, err)
			require.Equal(t, string(fail), string(bytes))
		})
	}
}

func TestEncode(t *testing.T) {
	bytes, err := newSecret(64)
	require.NoError(t, err)
	b64 := base64.StdEncoding.EncodeToString(bytes)
	input := hex.EncodeToString([]byte(b64))

	t.Run("validate", func(t *testing.T) {
		actual, err := EncodeDecode(testSource, input)
		require.NoError(t, err)
		require.Equal(t, input, actual)
	})
}

func TestNewPI(t *testing.T) {
	g, err := Read(testSource)
	require.NoError(t, err)
	_, err = newPaletteInfo(g)
	require.NoError(t, err)
}

// EncodeDecode reads a gif, encodes a random secret, then decodes it,
// returning the input and output secrets for comparison in a test context.
func EncodeDecode(path string, input string) (string, error) {
	g, err := Read(path)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}
	gNew, err := Encode(g, input)
	if err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}
	output, err := Decode(gNew)
	if err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return output, nil
}
