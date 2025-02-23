package cryptogif

import (
	"bufio"
	"crypto/rand"
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
func TestEncodePriorFails(t *testing.T) {
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
			t.Parallel()
			g, err := Read(testSource)
			require.NoError(t, err)
			g, err = Encode(g, []byte(fail))
			require.NoError(t, err)
			bytes, err := Decode(g)
			require.NoError(t, err)
			require.Equal(t, string(fail), string(bytes))
		})
	}
}

func TestEncode(t *testing.T) {
	testDataLens := []int{1, 2, 3, 4, 5}
	// testDataLens := []int{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144}
	// testDataLens := []int{128, 256, 512, 1024, 2048, 4096, 8192, 16384}
	for testNum, len := range testDataLens {
		t.Run(fmt.Sprintf("test-%d (%d)", testNum, len), func(t *testing.T) {
			t.Parallel()
			expect, actual, err := EncodeDecode(testSource, len)
			require.NoError(t, err)
			if string(expect) != string(actual) {
				t.Logf("\nExpect: %x\nActual: %x\n", expect, actual)
				// append to file
				f, err := os.OpenFile(testFailLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					t.Logf("open test case file: %v\n", err)
				}
				defer f.Close()
				if _, err := f.WriteString(fmt.Sprintf("%x %x\n", expect, actual)); err != nil {
					t.Logf("write test case file: %v\n", err)
				}
			}
			require.Equal(t, string(expect), string(actual))
		})
	}
}

func TestNewPI(t *testing.T) {
	g, err := Read(testSource)
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
