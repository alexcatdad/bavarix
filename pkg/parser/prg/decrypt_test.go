package prg

import (
	"bytes"
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

func TestXORDecryptRoundTrip(t *testing.T) {
	input := []byte{0xAA, 0xBB, 0xCC}
	result := XORDecrypt(XORDecrypt(input))
	assert.Equal(t, input, result, "XOR decrypt should be its own inverse")
}

func TestXORDecryptKnownValue(t *testing.T) {
	// 'A' (0x41) XOR 0xF7 = 0xB6
	input := []byte{0xB6}
	result := XORDecrypt(input)
	assert.Equal(t, []byte{'A'}, result)
}

func TestReadHeaderValid(t *testing.T) {
	data := make([]byte, 20)
	copy(data[:16], MagicHeader)
	data[16] = 0x01 // version = 1

	h, err := ReadHeader(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, MagicHeader, string(h.Magic[:]))
	assert.Equal(t, uint32(1), h.Version)
}

func TestReadHeaderInvalid(t *testing.T) {
	data := []byte("NOT AN EDIABAS FILE!")
	_, err := ReadHeader(bytes.NewReader(data))
	assert.ErrorIs(t, err, ErrInvalidMagic)
}

func TestReadHeaderTooShort(t *testing.T) {
	data := []byte("SHORT")
	_, err := ReadHeader(bytes.NewReader(data))
	assert.Error(t, err)
}
