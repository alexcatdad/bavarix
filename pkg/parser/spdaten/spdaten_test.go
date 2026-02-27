package spdaten

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "spdaten", name)
}

func TestParseGM5C05(t *testing.T) {
	path := testdataPath("GM5.C05")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test SP-Daten file not available")
	}

	module, err := Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "GM5.C05", module.Name)
}

func TestParseExtractsBlocks(t *testing.T) {
	path := testdataPath("GM5.C05")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test SP-Daten file not available")
	}

	module, err := Parse(path)
	require.NoError(t, err)

	assert.NotEmpty(t, module.CodingBlocks, "should contain coding data blocks")
}

func TestParseExtractsFields(t *testing.T) {
	path := testdataPath("GM5.C05")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test SP-Daten file not available")
	}

	module, err := Parse(path)
	require.NoError(t, err)

	hasLabels := false
	for _, block := range module.CodingBlocks {
		if len(block.Fields) > 0 {
			hasLabels = true
			break
		}
	}
	assert.True(t, hasLabels, "coding blocks should contain field definitions")
}

func TestParseInvalidFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "bad_spdaten_*")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	tmp.Write([]byte("not an SP-Daten file"))
	tmp.Close()

	_, err = Parse(tmp.Name())
	assert.Error(t, err)
}
