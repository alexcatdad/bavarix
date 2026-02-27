package prg

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

const (
	// jobTablePtrOffset is the offset in the PRG header where the pointer
	// to the job table is stored.
	jobTablePtrOffset = 0x88

	// jobEntrySize is the size of each entry in the job table:
	// 4 bytes (code offset, XOR encrypted) + 64 bytes (job name, XOR encrypted).
	jobEntrySize = 68

	// jobNameLen is the length of the job name field in each entry.
	jobNameLen = 64

	// minHeaderSize is the minimum valid PRG file size (magic + version + padding + pointers).
	minHeaderSize = 0x8C
)

// Job represents a named job extracted from a PRG file.
type Job struct {
	Name       string `json:"name"`
	CodeOffset uint32 `json:"code_offset"`
}

// ExtractJobs reads a PRG file and extracts all job names from the job table.
func ExtractJobs(path string) ([]Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("prg: reading file: %w", err)
	}

	if len(data) < 16 || string(data[:16]) != MagicHeader {
		return nil, ErrInvalidMagic
	}

	if len(data) < minHeaderSize {
		return nil, fmt.Errorf("prg: file too small for header (%d bytes)", len(data))
	}

	return extractJobsFromTable(data)
}

// extractJobsFromTable parses the binary job table structure.
func extractJobsFromTable(data []byte) ([]Job, error) {
	// Read the job table pointer from the header.
	tableOffset := binary.LittleEndian.Uint32(data[jobTablePtrOffset : jobTablePtrOffset+4])

	if tableOffset == 0 || tableOffset == 0xFFFFFFFF {
		// No job table present.
		return nil, nil
	}

	if int(tableOffset)+jobEntrySize > len(data) {
		return nil, fmt.Errorf("prg: job table offset 0x%X beyond file size %d", tableOffset, len(data))
	}

	// The first entry in the job table is a header entry.
	// Its first 4 bytes (raw, NOT encrypted) contain the total entry count
	// (including the header entry itself).
	totalEntries := int(binary.LittleEndian.Uint32(data[tableOffset : tableOffset+4]))

	if totalEntries <= 1 {
		// Only the header entry or no entries at all.
		return nil, nil
	}

	// The actual job entries start after the header entry.
	jobCount := totalEntries - 1
	startOffset := int(tableOffset) + jobEntrySize

	if startOffset+jobCount*jobEntrySize > len(data) {
		// Clamp to available data.
		jobCount = (len(data) - startOffset) / jobEntrySize
		if jobCount <= 0 {
			return nil, nil
		}
	}

	jobs := make([]Job, 0, jobCount)
	for i := 0; i < jobCount; i++ {
		entryOff := startOffset + i*jobEntrySize
		// The code offset is XOR encrypted like the name.
		rawOffset := binary.LittleEndian.Uint32(data[entryOff : entryOff+4])
		codeOffset := rawOffset ^ 0xF7F7F7F7
		encryptedName := data[entryOff+4 : entryOff+4+jobNameLen]
		decryptedName := XORDecrypt(encryptedName)

		// Extract the null-terminated string.
		name := extractNullTerminated(decryptedName)
		if name == "" {
			continue
		}

		jobs = append(jobs, Job{
			Name:       name,
			CodeOffset: codeOffset,
		})
	}

	return jobs, nil
}

// extractNullTerminated extracts a null-terminated string from a byte slice.
func extractNullTerminated(data []byte) string {
	idx := 0
	for idx < len(data) && data[idx] != 0 {
		idx++
	}
	return strings.TrimSpace(string(data[:idx]))
}
