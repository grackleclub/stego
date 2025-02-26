package cryptogif

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	_, err := read(testSource)
	require.NoError(t, err)
}

func TestWrite(t *testing.T) {
	g, err := read(testSource)
	require.NoError(t, err)
	err = write(g, testWrite)
	require.NoError(t, err)
}
