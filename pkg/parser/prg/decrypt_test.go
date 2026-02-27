package prg

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
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "prg", name)
}

func TestXORDecrypt(t *testing.T) {
	input := []byte{0xAA, 0xBB, 0xCC}
	encrypted := XORDecrypt(input)
	decrypted := XORDecrypt(encrypted)
	assert.Equal(t, input, decrypted)
}

func TestXORDecryptKnownValue(t *testing.T) {
	assert.Equal(t, []byte{0x00}, XORDecrypt([]byte{0xF7}))
	assert.Equal(t, []byte{0xF7}, XORDecrypt([]byte{0x00}))
}

func TestMagicHeader(t *testing.T) {
	assert.Equal(t, "@EDIABAS OBJECT\x00", MagicHeader)
}

func TestReadPRGFileHeader(t *testing.T) {
	path := testdataPath("C_GM5.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	header, err := ReadHeader(f)
	require.NoError(t, err)
	assert.Equal(t, MagicHeader, string(header.Magic[:]))
}

func TestDecryptPRGBody(t *testing.T) {
	path := testdataPath("C_GM5.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Decrypt a region after the header and look for readable ASCII
	body := data[0x90:]
	decrypted := XORDecrypt(body)

	found := false
	for i := 0; i < len(decrypted)-5; i++ {
		if isASCIIPrintable(decrypted[i : i+5]) {
			found = true
			break
		}
	}
	assert.True(t, found, "decrypted PRG body should contain readable ASCII strings")
}

func isASCIIPrintable(b []byte) bool {
	for _, c := range b {
		if c < 0x20 || c > 0x7E {
			return false
		}
	}
	return true
}
