package kwp2000

import (
	"errors"
	"fmt"
)

var (
	ErrBadChecksum    = errors.New("kwp2000: checksum mismatch")
	ErrFrameTooShort  = errors.New("kwp2000: frame too short")
	ErrLengthMismatch = errors.New("kwp2000: frame length mismatch")
)

type Frame struct {
	Target byte
	Source byte
	Data   []byte
}

func Checksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

func BuildFrame(target, source byte, data []byte) []byte {
	var frame []byte

	if len(data) <= 63 {
		header := byte(0x80) | byte(len(data))
		frame = make([]byte, 0, 3+len(data)+1)
		frame = append(frame, header, target, source)
	} else {
		frame = make([]byte, 0, 4+len(data)+1)
		frame = append(frame, 0x80, target, source, byte(len(data)))
	}

	frame = append(frame, data...)
	frame = append(frame, Checksum(frame))
	return frame
}

func ParseFrame(raw []byte) (Frame, error) {
	if len(raw) < 4 {
		return Frame{}, ErrFrameTooShort
	}

	payload := raw[:len(raw)-1]
	if Checksum(payload) != raw[len(raw)-1] {
		return Frame{}, fmt.Errorf("%w: expected 0x%02X, got 0x%02X",
			ErrBadChecksum, Checksum(payload), raw[len(raw)-1])
	}

	header := raw[0]
	target := raw[1]
	source := raw[2]

	dataLen := int(header & 0x3F)
	dataStart := 3

	if dataLen == 0 {
		if len(raw) < 5 {
			return Frame{}, ErrFrameTooShort
		}
		dataLen = int(raw[3])
		dataStart = 4
	}

	expectedTotal := dataStart + dataLen + 1
	if len(raw) != expectedTotal {
		return Frame{}, fmt.Errorf("%w: expected %d bytes, got %d",
			ErrLengthMismatch, expectedTotal, len(raw))
	}

	data := make([]byte, dataLen)
	copy(data, raw[dataStart:dataStart+dataLen])

	return Frame{
		Target: target,
		Source: source,
		Data:   data,
	}, nil
}
