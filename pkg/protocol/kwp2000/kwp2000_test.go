package kwp2000

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumAdditive(t *testing.T) {
	// 0x83 + 0x01 + 0xF1 + 0x19 + 0x02 + 0x0C = 0x19C, truncated to 0x9C
	data := []byte{0x83, 0x01, 0xF1, 0x19, 0x02, 0x0C}
	assert.Equal(t, byte(0x9C), Checksum(data))
}

func TestBuildFrameShort(t *testing.T) {
	frame := BuildFrame(0x01, 0xF1, []byte{0x19, 0x02, 0x0C})
	expected := []byte{0x83, 0x01, 0xF1, 0x19, 0x02, 0x0C, 0x9C}
	assert.Equal(t, expected, frame)
}

func TestBuildFrameExtended(t *testing.T) {
	data := make([]byte, 64)
	for i := range data {
		data[i] = byte(i)
	}
	frame := BuildFrame(0x01, 0xF1, data)
	assert.Equal(t, byte(0x80), frame[0])
	assert.Equal(t, byte(0x01), frame[1])
	assert.Equal(t, byte(0xF1), frame[2])
	assert.Equal(t, byte(64), frame[3])
	assert.Equal(t, 4+64+1, len(frame))
}

func TestParseFrameShort(t *testing.T) {
	raw := []byte{0x83, 0x01, 0xF1, 0x19, 0x02, 0x0C, 0x9C}
	frame, err := ParseFrame(raw)
	require.NoError(t, err)
	assert.Equal(t, byte(0x01), frame.Target)
	assert.Equal(t, byte(0xF1), frame.Source)
	assert.Equal(t, []byte{0x19, 0x02, 0x0C}, frame.Data)
}

func TestParseFrameBadChecksum(t *testing.T) {
	// Correct checksum for this payload is 0x9C, so 0xFF triggers mismatch
	raw := []byte{0x83, 0x01, 0xF1, 0x19, 0x02, 0x0C, 0xFF}
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrBadChecksum)
}

func TestParseFrameTooShort(t *testing.T) {
	raw := []byte{0x83, 0x01}
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrFrameTooShort)
}

func TestRoundTrip(t *testing.T) {
	original := []byte{0x19, 0x02, 0x0C}
	frame := BuildFrame(0x01, 0xF1, original)
	parsed, err := ParseFrame(frame)
	require.NoError(t, err)
	assert.Equal(t, byte(0x01), parsed.Target)
	assert.Equal(t, byte(0xF1), parsed.Source)
	assert.Equal(t, original, parsed.Data)
}
