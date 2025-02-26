package cryptogif

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaletteInfo(t *testing.T) {
	g, err := read(testSource)
	require.NoError(t, err)

	p, err := newPaletteInfo(g)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Greater(t, len(p), 0)

	t.Logf("palette info: %v found\n", len(p))
}

func TestSortByTone(t *testing.T) {
	g, err := read(testSource)
	require.NoError(t, err)

	pi, err := newPaletteInfo(g)
	require.NoError(t, err)
	require.NotNil(t, pi)

	sorted := sortByTone(pi[0])
	require.NotNil(t, sorted)
}
