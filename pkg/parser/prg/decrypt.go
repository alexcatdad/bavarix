package prg

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	MagicHeader = "@EDIABAS OBJECT\x00"
	xorKey      = 0xF7
)

var (
	ErrInvalidMagic = errors.New("prg: invalid magic header — not an EDIABAS PRG file")
)

type Header struct {
	Magic   [16]byte
	Version uint32
}

func XORDecrypt(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b ^ xorKey
	}
	return result
}

func ReadHeader(r io.Reader) (Header, error) {
	var h Header
	if err := binary.Read(r, binary.LittleEndian, &h.Magic); err != nil {
		return Header{}, fmt.Errorf("prg: reading magic: %w", err)
	}

	if string(h.Magic[:]) != MagicHeader {
		return Header{}, ErrInvalidMagic
	}

	if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
		return Header{}, fmt.Errorf("prg: reading version: %w", err)
	}

	return h, nil
}
