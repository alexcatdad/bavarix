package vault

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempVault(t *testing.T) *Vault {
	t.Helper()
	tmp, err := os.CreateTemp("", "vault_test_*.db")
	require.NoError(t, err)
	tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	v, err := Open(tmp.Name())
	require.NoError(t, err)
	t.Cleanup(func() { v.Close() })

	return v
}

func TestOpenAndClose(t *testing.T) {
	v := tempVault(t)
	assert.NotNil(t, v)
}

func TestSaveAndRetrieve(t *testing.T) {
	v := tempVault(t)

	coding := []byte{0x01, 0x02, 0x03, 0x04}
	err := v.Save("E46", "GM5", "C05", coding)
	require.NoError(t, err)

	entries, err := v.List("E46", "GM5")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, coding, entries[0].Data)
	assert.Equal(t, "GM5", entries[0].Module)
	assert.Equal(t, "C05", entries[0].Version)
}

func TestSaveMultipleVersions(t *testing.T) {
	v := tempVault(t)

	err := v.Save("E46", "GM5", "C05", []byte{0x01})
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	err = v.Save("E46", "GM5", "C05", []byte{0x02})
	require.NoError(t, err)

	entries, err := v.List("E46", "GM5")
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	assert.Equal(t, []byte{0x02}, entries[0].Data)
	assert.Equal(t, []byte{0x01}, entries[1].Data)
}

func TestGetLatest(t *testing.T) {
	v := tempVault(t)

	v.Save("E46", "GM5", "C05", []byte{0x01})
	time.Sleep(10 * time.Millisecond)
	v.Save("E46", "GM5", "C05", []byte{0x02})

	entry, err := v.Latest("E46", "GM5")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x02}, entry.Data)
}

func TestGetLatestEmpty(t *testing.T) {
	v := tempVault(t)

	_, err := v.Latest("E46", "GM5")
	assert.ErrorIs(t, err, ErrNoBackup)
}

func TestListDifferentModules(t *testing.T) {
	v := tempVault(t)

	v.Save("E46", "GM5", "C05", []byte{0x01})
	v.Save("E46", "KMB", "C06", []byte{0x02})

	gm5, _ := v.List("E46", "GM5")
	kmb, _ := v.List("E46", "KMB")

	assert.Len(t, gm5, 1)
	assert.Len(t, kmb, 1)
}
