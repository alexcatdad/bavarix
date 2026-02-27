package ds2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumXOR(t *testing.T) {
	data := []byte{0x12, 0x04, 0x00}
	assert.Equal(t, byte(0x16), Checksum(data))
}

func TestChecksumEmpty(t *testing.T) {
	data := []byte{}
	assert.Equal(t, byte(0x00), Checksum(data))
}

func TestChecksumSingleByte(t *testing.T) {
	data := []byte{0xAB}
	assert.Equal(t, byte(0xAB), Checksum(data))
}

func TestBuildFrame(t *testing.T) {
	frame := BuildFrame(0x12, []byte{0x00})
	assert.Equal(t, []byte{0x12, 0x04, 0x00, 0x16}, frame)
}

func TestBuildFrameNoData(t *testing.T) {
	frame := BuildFrame(0x12, []byte{})
	assert.Equal(t, []byte{0x12, 0x03, 0x11}, frame)
}

func TestParseFrame(t *testing.T) {
	raw := []byte{0x12, 0x04, 0x00, 0x16}
	frame, err := ParseFrame(raw)
	require.NoError(t, err)
	assert.Equal(t, byte(0x12), frame.Address)
	assert.Equal(t, []byte{0x00}, frame.Data)
}

func TestParseFrameBadChecksum(t *testing.T) {
	raw := []byte{0x12, 0x04, 0x00, 0xFF}
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrBadChecksum)
}

func TestParseFrameTooShort(t *testing.T) {
	raw := []byte{0x12}
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrFrameTooShort)
}

func TestParseFrameLengthMismatch(t *testing.T) {
	raw := []byte{0x12, 0x09, 0x00, 0x16}
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrLengthMismatch)
}

func TestRoundTrip(t *testing.T) {
	original := []byte{0xA0, 0x88, 0x36, 0x94, 0x83}
	frame := BuildFrame(0x00, original)
	parsed, err := ParseFrame(frame)
	require.NoError(t, err)
	assert.Equal(t, byte(0x00), parsed.Address)
	assert.Equal(t, original, parsed.Data)
}
