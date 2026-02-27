package translations

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "translations", name)
}

func TestLoadTranslations(t *testing.T) {
	path := testdataPath("sample.csv")
	dict, err := Load(path)
	require.NoError(t, err)
	assert.NotEmpty(t, dict.Entries)
}

func TestTranslateKnownValue(t *testing.T) {
	path := testdataPath("sample.csv")
	dict, err := Load(path)
	require.NoError(t, err)

	result, found := dict.Translate("0.7_sekunden")
	assert.True(t, found)
	assert.Equal(t, "0.7 seconds", result)
}

func TestTranslateUnknownValue(t *testing.T) {
	path := testdataPath("sample.csv")
	dict, err := Load(path)
	require.NoError(t, err)

	_, found := dict.Translate("nonexistent_key_xyz")
	assert.False(t, found)
}

func TestTranslateCaseInsensitive(t *testing.T) {
	path := testdataPath("sample.csv")
	dict, err := Load(path)
	require.NoError(t, err)

	result1, found1 := dict.Translate("0.7_sekunden")
	result2, found2 := dict.Translate("0.7_SEKUNDEN")
	assert.Equal(t, found1, found2)
	if found1 {
		assert.Equal(t, result1, result2)
	}
}

func TestSkipsMetadataRows(t *testing.T) {
	path := testdataPath("sample.csv")
	dict, err := Load(path)
	require.NoError(t, err)

	_, found := dict.Translate("CONTRIBUTORS")
	assert.False(t, found)
}

func TestLoadInvalidPath(t *testing.T) {
	_, err := Load("/nonexistent/path.csv")
	assert.Error(t, err)
}
