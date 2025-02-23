package cryptogif

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
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
	// secret, err := newSecret(10)
	// require.NoError(t, err)

	secret := []byte("hello-world")

	_, err := encode(secret, path.Join("img", "earth.gif"), path.Join("img", "earth_out.gif"))
	require.NoError(t, err)

	result, err := decode(path.Join("img", "earth_out.gif"))
	require.NoError(t, err)

	t.Logf("original: %v", string(secret))
	t.Logf("result: %v", string(result))

	require.Equal(t, len(secret), len(result))
}

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
	t.Log("original: ", secret)
	t.Log("crushed: ", crushed)
	t.Log("stretched: ", stretched)
	require.Equal(t, len(secret)*2, len(crushed))
}

func TestNewPI(t *testing.T) {
	g, err := Read(path.Join("img", "wiki.gif"))
	require.NoError(t, err)

	pi, err := newPI(g)
	require.NoError(t, err)
	t.Logf("pi: %v\n", pi)
}
