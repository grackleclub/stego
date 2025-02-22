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
	_, err := setFloor(path.Join("img", "wiki.gif"), path.Join("img", "wiki_out.gif"))
	require.NoError(t, err)
}
