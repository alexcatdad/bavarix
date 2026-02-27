package prg

import (
	"fmt"
	"os"
	"strings"
)

// jobRecordSize is the fixed size of each job table entry:
// 4 bytes for the code address + 64 bytes for the null-terminated name.
const jobRecordSize = 68

// jobNameMaxLen is the maximum length of a job name field within a record.
const jobNameMaxLen = 64

// headerJobTableOffset is the position in the PRG header where the
// job table start offset is stored as a uint32 LE value.
const headerJobTableOffset = 0x88

// Job represents a single EDIABAS diagnostic job extracted from a PRG file.
type Job struct {
	Name    string
	Address uint32
}

// ExtractJobs reads a PRG file and returns all diagnostic job definitions
// found in the job table.
//
// The PRG file format stores jobs in a fixed-size record table. The table
// location is specified by a uint32 LE offset at position 0x88 in the file
// header. The first 4 bytes at the table start are the job count stored as
// an unencrypted uint32 LE in the raw file. The table body is XOR-encrypted
// with key 0xF7. Each 68-byte record contains a 4-byte code address followed
// by a 64-byte null-terminated job name. The first record's address field
// (after decryption) is not a valid code address since those raw bytes held
// the job count.
func ExtractJobs(path string) ([]Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("prg: reading file: %w", err)
	}

	if len(data) < 16 || string(data[:16]) != MagicHeader {
		return nil, ErrInvalidMagic
	}

	if len(data) < headerJobTableOffset+4 {
		return nil, fmt.Errorf("prg: file too short for header offset table")
	}

	jobTableStart := int(leU32(data, headerJobTableOffset))
	if jobTableStart >= len(data) {
		return nil, fmt.Errorf("prg: job table offset 0x%X beyond file size 0x%X", jobTableStart, len(data))
	}

	// The first 4 bytes at the job table start contain the job count
	// as an unencrypted uint32 LE value in the raw file data.
	if jobTableStart+4 > len(data) {
		return nil, fmt.Errorf("prg: job table too short")
	}
	jobCount := int(leU32(data, jobTableStart))

	// Decrypt the entire job table region for reading names and addresses.
	decrypted := XORDecrypt(data[jobTableStart:])

	if jobCount <= 0 || jobCount > 10000 {
		return nil, fmt.Errorf("prg: unreasonable job count %d", jobCount)
	}

	requiredBytes := jobCount * jobRecordSize
	if requiredBytes > len(decrypted) {
		return nil, fmt.Errorf("prg: job table claims %d records but only %d bytes available",
			jobCount, len(decrypted))
	}

	jobs := make([]Job, 0, jobCount)
	for i := 0; i < jobCount; i++ {
		base := i * jobRecordSize
		addr := leU32(decrypted, base)
		name := extractNullTerminated(decrypted[base+4 : base+jobRecordSize])

		// Skip records with empty names (should not happen in well-formed files).
		if name == "" {
			continue
		}

		jobs = append(jobs, Job{
			Name:    name,
			Address: addr,
		})
	}

	return jobs, nil
}

// leU32 reads a little-endian uint32 from data at the given offset.
func leU32(data []byte, off int) uint32 {
	return uint32(data[off]) |
		uint32(data[off+1])<<8 |
		uint32(data[off+2])<<16 |
		uint32(data[off+3])<<24
}

// extractNullTerminated extracts a null-terminated string from a byte slice,
// trimming any trailing whitespace.
func extractNullTerminated(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return strings.TrimSpace(string(b[:i]))
		}
	}
	return strings.TrimSpace(string(b))
}
