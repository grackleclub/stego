package cryptogif

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPaletteInfo(t *testing.T) {
	g, err := Read(testSource)
	require.NoError(t, err)

	pi, err := newPaletteInfo(g)
	require.NoError(t, err)
	require.NotNil(t, pi)

}

func TestSortByTone(t *testing.T) {
	g, err := Read(testSource)
	require.NoError(t, err)

	pi, err := newPaletteInfo(g)
	require.NoError(t, err)
	require.NotNil(t, pi)

	sorted := sortByTone(pi[0])
	require.NotNil(t, sorted)

}
