package ds2

import (
	"errors"
	"fmt"
)

var (
	ErrBadChecksum    = errors.New("ds2: checksum mismatch")
	ErrFrameTooShort  = errors.New("ds2: frame too short (minimum 3 bytes)")
	ErrLengthMismatch = errors.New("ds2: frame length does not match length byte")
)

type Frame struct {
	Address byte
	Data    []byte
}

func Checksum(data []byte) byte {
	var cs byte
	for _, b := range data {
		cs ^= b
	}
	return cs
}

func BuildFrame(address byte, data []byte) []byte {
	length := byte(len(data) + 3)
	frame := make([]byte, 0, int(length))
	frame = append(frame, address, length)
	frame = append(frame, data...)
	frame = append(frame, Checksum(frame))
	return frame
}

func ParseFrame(raw []byte) (Frame, error) {
	if len(raw) < 3 {
		return Frame{}, ErrFrameTooShort
	}

	length := int(raw[1])
	if length != len(raw) {
		return Frame{}, fmt.Errorf("%w: expected %d, got %d", ErrLengthMismatch, length, len(raw))
	}

	payload := raw[:len(raw)-1]
	if Checksum(payload) != raw[len(raw)-1] {
		return Frame{}, fmt.Errorf("%w: expected 0x%02X, got 0x%02X", ErrBadChecksum, Checksum(payload), raw[len(raw)-1])
	}

	data := make([]byte, len(raw)-3)
	copy(data, raw[2:len(raw)-1])

	return Frame{
		Address: raw[0],
		Data:    data,
	}, nil
}
