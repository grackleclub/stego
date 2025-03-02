package stego

import (
	"testing"

	"github.com/stretchr/testify/require"
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
