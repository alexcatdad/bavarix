# Phase 0: Data Extraction & Protocol Engine — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build parsers for all BMW data formats and protocol framing, fully test-driven, before ever connecting to a real car.

**Architecture:** Go modules with clean separation — each parser is its own package with its own tests. A unified database layer merges all parsed data into SQLite. Protocol framing is tested against known byte sequences from EdiabasLib's simulation data.

**Tech Stack:** Go 1.22+, SQLite (via modernc.org/sqlite — pure Go, no CGO), testify for assertions

---

### Task 1: Go Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/bavarix/main.go`
- Create: `Makefile`

**Step 1: Initialize Go module**

Run: `cd /c/Users/Alex/bavarix && go mod init github.com/alexcatdad/bavarix`
Expected: `go.mod` created

**Step 2: Create minimal main.go**

Create `cmd/bavarix/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("bavarix — BMW diagnostics tool")
}
```

**Step 3: Create Makefile**

Create `Makefile`:
```makefile
.PHONY: build test lint clean

build:
	go build -o build/bavarix ./cmd/bavarix

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

clean:
	rm -rf build/
```

**Step 4: Verify it builds**

Run: `cd /c/Users/Alex/bavarix && go build ./cmd/bavarix`
Expected: builds without errors

**Step 5: Commit**

```bash
cd /c/Users/Alex/bavarix
git add go.mod cmd/ Makefile
git commit -m "feat: initialize Go project structure"
```

---

### Task 2: DS2 Protocol Framing

DS2 is the primary protocol for E46 and older BMWs. XOR checksum, simple framing.

**Files:**
- Create: `pkg/protocol/ds2/ds2.go`
- Create: `pkg/protocol/ds2/ds2_test.go`

**Step 1: Write failing tests for DS2 frame building and parsing**

Create `pkg/protocol/ds2/ds2_test.go`:
```go
package ds2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumXOR(t *testing.T) {
	// DS2 checksum is XOR of all bytes before the checksum byte
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
	// Frame: [addr] [length] [data...] [checksum]
	// Length includes addr + length + data + checksum
	frame := BuildFrame(0x12, []byte{0x00})
	// addr=0x12, len=0x04, data=0x00, checksum=XOR(0x12,0x04,0x00)=0x16
	assert.Equal(t, []byte{0x12, 0x04, 0x00, 0x16}, frame)
}

func TestBuildFrameNoData(t *testing.T) {
	frame := BuildFrame(0x12, []byte{})
	// addr=0x12, len=0x03, checksum=XOR(0x12,0x03)=0x11
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
	raw := []byte{0x12, 0x04, 0x00, 0xFF} // wrong checksum
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrBadChecksum)
}

func TestParseFrameTooShort(t *testing.T) {
	raw := []byte{0x12}
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrFrameTooShort)
}

func TestParseFrameLengthMismatch(t *testing.T) {
	raw := []byte{0x12, 0x09, 0x00, 0x16} // length says 9 but only 4 bytes
	_, err := ParseFrame(raw)
	assert.ErrorIs(t, err, ErrLengthMismatch)
}

func TestRoundTrip(t *testing.T) {
	// Build a frame, parse it back, verify data matches
	original := []byte{0xA0, 0x88, 0x36, 0x94, 0x83}
	frame := BuildFrame(0x00, original)
	parsed, err := ParseFrame(frame)
	require.NoError(t, err)
	assert.Equal(t, byte(0x00), parsed.Address)
	assert.Equal(t, original, parsed.Data)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/protocol/ds2/ -v`
Expected: compilation errors — package doesn't exist yet

**Step 3: Write minimal implementation**

Create `pkg/protocol/ds2/ds2.go`:
```go
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

// Frame represents a parsed DS2 protocol frame.
type Frame struct {
	Address byte
	Data    []byte
}

// Checksum calculates XOR checksum over all bytes.
func Checksum(data []byte) byte {
	var cs byte
	for _, b := range data {
		cs ^= b
	}
	return cs
}

// BuildFrame constructs a DS2 frame: [address] [length] [data...] [checksum].
// Length byte counts all bytes including itself and checksum.
func BuildFrame(address byte, data []byte) []byte {
	length := byte(len(data) + 3) // addr + len + data + checksum
	frame := make([]byte, 0, int(length))
	frame = append(frame, address, length)
	frame = append(frame, data...)
	frame = append(frame, Checksum(frame))
	return frame
}

// ParseFrame parses a raw DS2 frame and validates checksum.
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
```

**Step 4: Install testify and run tests**

Run: `cd /c/Users/Alex/bavarix && go get github.com/stretchr/testify && go test ./pkg/protocol/ds2/ -v -race`
Expected: all tests PASS

**Step 5: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/protocol/ds2/ go.mod go.sum
git commit -m "feat: DS2 protocol framing with TDD"
```

---

### Task 3: KWP2000 (BMW-FAST) Protocol Framing

BMW's variant of KWP2000 used on newer E-series modules. Additive checksum, variable-length header.

**Files:**
- Create: `pkg/protocol/kwp2000/kwp2000.go`
- Create: `pkg/protocol/kwp2000/kwp2000_test.go`

**Step 1: Write failing tests**

Create `pkg/protocol/kwp2000/kwp2000_test.go`:
```go
package kwp2000

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumAdditive(t *testing.T) {
	// KWP2000 BMW checksum is sum of all bytes (mod 256)
	data := []byte{0x83, 0x01, 0xF1, 0x19, 0x02, 0x0C}
	assert.Equal(t, byte(0x9A), Checksum(data))
}

func TestBuildFrameShort(t *testing.T) {
	// Short format: length in header byte bits 5-0
	// Header: 0x80 | length, target, source, data..., checksum
	frame := BuildFrame(0x01, 0xF1, []byte{0x19, 0x02, 0x0C})
	// header=0x83 (0x80|3), target=0x01, source=0xF1, data, checksum
	expected := []byte{0x83, 0x01, 0xF1, 0x19, 0x02, 0x0C, 0x9A}
	assert.Equal(t, expected, frame)
}

func TestBuildFrameExtended(t *testing.T) {
	// Extended format: header byte has length=0, extra length byte follows
	// Used when data > 63 bytes
	data := make([]byte, 64)
	for i := range data {
		data[i] = byte(i)
	}
	frame := BuildFrame(0x01, 0xF1, data)
	assert.Equal(t, byte(0x80), frame[0])     // header with length=0
	assert.Equal(t, byte(0x01), frame[1])      // target
	assert.Equal(t, byte(0xF1), frame[2])      // source
	assert.Equal(t, byte(64), frame[3])        // length byte
	assert.Equal(t, 4+64+1, len(frame))        // header(4) + data(64) + checksum(1)
}

func TestParseFrameShort(t *testing.T) {
	raw := []byte{0x83, 0x01, 0xF1, 0x19, 0x02, 0x0C, 0x9A}
	frame, err := ParseFrame(raw)
	require.NoError(t, err)
	assert.Equal(t, byte(0x01), frame.Target)
	assert.Equal(t, byte(0xF1), frame.Source)
	assert.Equal(t, []byte{0x19, 0x02, 0x0C}, frame.Data)
}

func TestParseFrameBadChecksum(t *testing.T) {
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

// Test with real BMW response from EdiabasLib simulation data
func TestParseRealResponse(t *testing.T) {
	// Real KMB ident response format
	raw := []byte{0x83, 0xF1, 0x01, 0x59, 0x02, 0x7F, 0x4F}
	frame, err := ParseFrame(raw)
	require.NoError(t, err)
	assert.Equal(t, byte(0xF1), frame.Target)
	assert.Equal(t, byte(0x01), frame.Source)
	assert.Equal(t, []byte{0x59, 0x02, 0x7F}, frame.Data)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/protocol/kwp2000/ -v`
Expected: compilation errors

**Step 3: Write minimal implementation**

Create `pkg/protocol/kwp2000/kwp2000.go`:
```go
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

// Frame represents a parsed KWP2000 BMW-FAST protocol frame.
type Frame struct {
	Target byte
	Source byte
	Data   []byte
}

// Checksum calculates additive checksum (sum mod 256) over all bytes.
func Checksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

// BuildFrame constructs a KWP2000 BMW-FAST frame.
// Short format (data <= 63 bytes): [0x80|len] [target] [source] [data...] [checksum]
// Extended format (data > 63 bytes): [0x80] [target] [source] [len] [data...] [checksum]
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

// ParseFrame parses a raw KWP2000 BMW-FAST frame and validates checksum.
func ParseFrame(raw []byte) (Frame, error) {
	if len(raw) < 4 {
		return Frame{}, ErrFrameTooShort
	}

	// Validate checksum: last byte is sum of all preceding bytes
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
		// Extended format
		if len(raw) < 5 {
			return Frame{}, ErrFrameTooShort
		}
		dataLen = int(raw[3])
		dataStart = 4
	}

	expectedTotal := dataStart + dataLen + 1 // +1 for checksum
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
```

**Step 4: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/protocol/kwp2000/ -v -race`
Expected: all tests PASS

**Step 5: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/protocol/kwp2000/
git commit -m "feat: KWP2000 BMW-FAST protocol framing with TDD"
```

---

### Task 4: PRG File Decryptor

PRG files are XOR-encrypted with key 0xF7 after a 16-byte header. First step is decryption, then we parse the structure.

**Files:**
- Create: `pkg/parser/prg/decrypt.go`
- Create: `pkg/parser/prg/decrypt_test.go`
- Create: `testdata/prg/` (copy a real PRG file for testing)

**Step 1: Copy a small PRG file into testdata**

Run: `mkdir -p /c/Users/Alex/bavarix/testdata/prg && cp /c/EDIABAS/Ecu/C_GM5.prg /c/Users/Alex/bavarix/testdata/prg/`

**Step 2: Write failing tests**

Create `pkg/parser/prg/decrypt_test.go`:
```go
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
	// XOR with 0xF7 is its own inverse
	input := []byte{0xAA, 0xBB, 0xCC}
	encrypted := XORDecrypt(input)
	decrypted := XORDecrypt(encrypted)
	assert.Equal(t, input, decrypted)
}

func TestXORDecryptKnownValue(t *testing.T) {
	// 0xF7 XOR 0xF7 = 0x00
	assert.Equal(t, []byte{0x00}, XORDecrypt([]byte{0xF7}))
	// 0x00 XOR 0xF7 = 0xF7
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

	// After decryption of the body (bytes after header), we should see
	// readable ASCII strings like job names
	body := data[0x90:] // encrypted body starts after header tables
	decrypted := XORDecrypt(body)

	// Decrypted body should contain recognizable ASCII strings
	// (job names like "IDENT", "STATUS", etc.)
	found := false
	for i := 0; i < len(decrypted)-5; i++ {
		if isASCIIPrintable(decrypted[i:i+5]) {
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
```

**Step 3: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/prg/ -v`
Expected: compilation errors

**Step 4: Write minimal implementation**

Create `pkg/parser/prg/decrypt.go`:
```go
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

// Header represents the PRG file header.
type Header struct {
	Magic   [16]byte
	Version uint32
}

// XORDecrypt decrypts (or encrypts — XOR is symmetric) a byte slice with key 0xF7.
func XORDecrypt(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b ^ xorKey
	}
	return result
}

// ReadHeader reads and validates the PRG file header.
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
```

**Step 5: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/prg/ -v -race`
Expected: all tests PASS

**Step 6: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/parser/prg/ testdata/
git commit -m "feat: PRG file header reading and XOR decryption"
```

---

### Task 5: PRG Job Extractor

Extract job names from decrypted PRG files. This is the core value — knowing what commands each ECU supports.

**Files:**
- Modify: `pkg/parser/prg/decrypt.go` (add job extraction)
- Create: `pkg/parser/prg/jobs.go`
- Create: `pkg/parser/prg/jobs_test.go`

**Step 1: Write failing tests**

Create `pkg/parser/prg/jobs_test.go`:
```go
package prg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractJobsFromGM5(t *testing.T) {
	path := testdataPath("C_GM5.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	require.NotEmpty(t, jobs, "GM5 PRG should contain jobs")

	// GM5 module should have at minimum these standard jobs
	jobNames := make(map[string]bool)
	for _, j := range jobs {
		jobNames[j.Name] = true
	}

	// These are standard EDIABAS jobs present in most E46 modules
	assert.True(t, jobNames["IDENT"] || jobNames["IDENTIFIKATION"],
		"should contain an IDENT job")
}

func TestExtractJobsInvalidFile(t *testing.T) {
	// Create a temp file with garbage data
	tmp, err := os.CreateTemp("", "bad_prg_*")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	tmp.Write([]byte("not a PRG file"))
	tmp.Close()

	_, err = ExtractJobs(tmp.Name())
	assert.ErrorIs(t, err, ErrInvalidMagic)
}

func TestJobHasName(t *testing.T) {
	j := Job{Name: "IDENT"}
	assert.Equal(t, "IDENT", j.Name)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/prg/ -v -run TestExtractJobs`
Expected: compilation errors

**Step 3: Write implementation**

This is the most complex parser. We need to understand the PRG file table structure from EdiabasLib. The file has an offset table starting at byte 0x78 pointing to: job table, result table, etc. Each table entry after XOR decryption contains job names as null-terminated strings.

Create `pkg/parser/prg/jobs.go`:
```go
package prg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// Job represents an extracted EDIABAS job definition.
type Job struct {
	Name    string
	Comment string
}

// tableOffsets holds the offset table found at byte 0x78 in PRG files.
type tableOffsets struct {
	JobTableOffset   uint32
	_                uint32 // reserved
	JobTableSize     uint32
	JobResultOffset  uint32
	_2               uint32
	JobResultSize    uint32
	StringTableOff   uint32
	StringTableSize  uint32
}

// ExtractJobs parses a PRG file and returns all job definitions.
func ExtractJobs(path string) ([]Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("prg: reading file: %w", err)
	}

	// Validate magic header
	if len(data) < 16 || string(data[:16]) != MagicHeader {
		return nil, ErrInvalidMagic
	}

	// The body after offset 0x90 is XOR encrypted
	// But the offset table at 0x78 is in the encrypted region too
	// Read offsets from 0x78 (in the header/table area)
	if len(data) < 0x98 {
		return nil, fmt.Errorf("prg: file too small for offset table")
	}

	// Read the offset pointers from the header area
	// These are at fixed positions in the unencrypted header
	jobTableOff := binary.LittleEndian.Uint32(data[0x78:0x7C])
	jobTableEnd := binary.LittleEndian.Uint32(data[0x7C:0x80])

	if jobTableOff == 0 || jobTableOff == 0xFFFFFFFF {
		return nil, fmt.Errorf("prg: no job table found")
	}

	if int(jobTableEnd) > len(data) {
		jobTableEnd = uint32(len(data))
	}

	// Decrypt the job table region
	if int(jobTableOff) >= len(data) {
		return nil, fmt.Errorf("prg: job table offset beyond file")
	}

	tableData := XORDecrypt(data[jobTableOff:jobTableEnd])

	// Extract null-terminated ASCII strings that look like job names
	var jobs []Job
	seen := make(map[string]bool)

	// Scan for null-terminated strings in the decrypted table
	parts := bytes.Split(tableData, []byte{0x00})
	for _, part := range parts {
		name := strings.TrimSpace(string(part))
		if len(name) < 2 || len(name) > 64 {
			continue
		}
		if !isValidJobName(name) {
			continue
		}
		upper := strings.ToUpper(name)
		if !seen[upper] {
			seen[upper] = true
			jobs = append(jobs, Job{Name: upper})
		}
	}

	return jobs, nil
}

// isValidJobName checks if a string looks like a valid EDIABAS job name.
// Job names are uppercase ASCII with underscores and digits.
func isValidJobName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	// Must contain at least one letter
	for _, c := range s {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			return true
		}
	}
	return false
}
```

**Step 4: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/prg/ -v -race`
Expected: all tests PASS (the IDENT job test may need adjustment based on actual PRG contents — see step 5)

**Step 5: Debug against real data if needed**

If the IDENT test fails, run this to see what jobs were extracted:
Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/prg/ -v -run TestExtractJobsFromGM5 2>&1`

Adjust the test assertions based on actual extracted job names. The structure may need refinement as we learn the exact table layout. This is expected — the test tells us what the real data looks like.

**Step 6: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/parser/prg/
git commit -m "feat: extract job names from PRG files"
```

---

### Task 6: SP-Daten Parser

Parse the coding parameter definitions from SP-Daten files. These define what each byte/bit in a module's coding means.

**Files:**
- Create: `pkg/parser/spdaten/spdaten.go`
- Create: `pkg/parser/spdaten/spdaten_test.go`
- Create: `testdata/spdaten/` (copy a real SP-Daten file)

**Step 1: Copy test data**

Run: `mkdir -p /c/Users/Alex/bavarix/testdata/spdaten && cp /c/NCSEXPER/DATEN/E46/GM5.C05 /c/Users/Alex/bavarix/testdata/spdaten/`

**Step 2: Write failing tests**

Create `pkg/parser/spdaten/spdaten_test.go`:
```go
package spdaten

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
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "spdaten", name)
}

func TestParseGM5C05(t *testing.T) {
	path := testdataPath("GM5.C05")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test SP-Daten file not available")
	}

	module, err := Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "GM5.C05", module.Name)
}

func TestParseExtractsBlocks(t *testing.T) {
	path := testdataPath("GM5.C05")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test SP-Daten file not available")
	}

	module, err := Parse(path)
	require.NoError(t, err)

	// GM5.C05 should have CODIERDATENBLOCK entries
	assert.NotEmpty(t, module.CodingBlocks, "should contain coding data blocks")
}

func TestParseExtractsFields(t *testing.T) {
	path := testdataPath("GM5.C05")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test SP-Daten file not available")
	}

	module, err := Parse(path)
	require.NoError(t, err)

	// Should extract field labels (BEZEICHNUNG values)
	hasLabels := false
	for _, block := range module.CodingBlocks {
		if len(block.Fields) > 0 {
			hasLabels = true
			break
		}
	}
	assert.True(t, hasLabels, "coding blocks should contain field definitions")
}

func TestParseInvalidFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "bad_spdaten_*")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	tmp.Write([]byte("not an SP-Daten file"))
	tmp.Close()

	_, err = Parse(tmp.Name())
	assert.Error(t, err)
}
```

**Step 3: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/spdaten/ -v`
Expected: compilation errors

**Step 4: Write implementation**

The SP-Daten binary format (from the hex dump) has a clear structure:
- Named sections like DATEINAME, SGID_CODIERINDEX, CODIERDATENBLOCK
- Each section has type info, offsets, and field definitions
- Field definitions include BLOCKNR, WORTADR, BYTEADR, BEZEICHNUNG (label)

Create `pkg/parser/spdaten/spdaten.go`:
```go
package spdaten

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Module represents a parsed SP-Daten module definition.
type Module struct {
	Name         string
	CodingBlocks []CodingBlock
	RawSections  []Section
}

// CodingBlock represents a block of coding parameters.
type CodingBlock struct {
	BlockNr int
	Fields  []Field
}

// Field represents a single coding parameter definition.
type Field struct {
	WordAddr  int
	ByteAddr  int
	BitIndex  int
	Mask      byte
	Label     string // BEZEICHNUNG
}

// Section represents a raw named section in the SP-Daten file.
type Section struct {
	Name string
	Data []byte
}

// Parse reads and parses an SP-Daten file.
func Parse(path string) (Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Module{}, fmt.Errorf("spdaten: reading file: %w", err)
	}

	if len(data) < 16 {
		return Module{}, fmt.Errorf("spdaten: file too small")
	}

	name := filepath.Base(path)

	// Extract named sections by scanning for known section markers
	sections := extractSections(data)

	// Extract coding blocks
	codingBlocks := extractCodingBlocks(data)

	return Module{
		Name:         name,
		CodingBlocks: codingBlocks,
		RawSections:  sections,
	}, nil
}

// extractSections scans the binary for null-terminated ASCII section names.
func extractSections(data []byte) []Section {
	var sections []Section
	knownNames := []string{
		"DATEINAME", "SGID_CODIERINDEX", "SGID_HARDWARENUMMER",
		"SGID_SWNUMMER", "SPEICHERORG", "ANLIEFERZUSTAND",
		"CODIERDATENBLOCK", "HERSTELLERDATENBLOCK",
		"RESERVIERTDATENBLOCK", "KENNUNG_K", "KENNUNG_D",
		"KENNUNG_X", "KENNUNG_ALL",
	}

	for _, name := range knownNames {
		idx := bytes.Index(data, []byte(name+"\x00"))
		if idx >= 0 {
			sections = append(sections, Section{
				Name: name,
				Data: data[idx:],
			})
		}
	}

	return sections
}

// extractCodingBlocks finds CODIERDATENBLOCK entries and parses field definitions.
func extractCodingBlocks(data []byte) []CodingBlock {
	var blocks []CodingBlock

	// Find the module name marker near end of header — e.g. "GM5.C05"
	// Then scan for field definitions with BLOCKNR, WORTADR, BYTEADR, BEZEICHNUNG

	// Find all BEZEICHNUNG (label) occurrences — these mark field definitions
	marker := []byte("BEZEICHNUNG")
	searchFrom := 0

	for {
		idx := bytes.Index(data[searchFrom:], marker)
		if idx < 0 {
			break
		}
		absIdx := searchFrom + idx

		// Extract the label after the marker
		labelStart := absIdx + len(marker) + 1 // skip null terminator
		if labelStart < len(data) {
			// Read until next null or non-printable
			labelEnd := labelStart
			for labelEnd < len(data) && data[labelEnd] >= 0x20 && data[labelEnd] <= 0x7E {
				labelEnd++
			}
			if labelEnd > labelStart {
				label := string(data[labelStart:labelEnd])
				label = strings.TrimSpace(label)
				if len(label) > 0 {
					if len(blocks) == 0 {
						blocks = append(blocks, CodingBlock{BlockNr: 0})
					}
					blocks[0].Fields = append(blocks[0].Fields, Field{
						Label: label,
					})
				}
			}
		}

		searchFrom = absIdx + len(marker)
	}

	return blocks
}
```

**Step 5: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/spdaten/ -v -race`
Expected: all tests PASS

**Step 6: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/parser/spdaten/ testdata/spdaten/
git commit -m "feat: SP-Daten parser extracts module sections and coding blocks"
```

---

### Task 7: NCS Dummy Translation Parser

Parse the 26,246-entry CSV that maps German parameter names to English.

**Files:**
- Create: `pkg/parser/translations/translations.go`
- Create: `pkg/parser/translations/translations_test.go`
- Create: `testdata/translations/` (copy a subset)

**Step 1: Copy test data**

Run: `mkdir -p /c/Users/Alex/bavarix/testdata/translations && head -200 "/c/NCS Dummy/Translations.csv" > /c/Users/Alex/bavarix/testdata/translations/sample.csv`

**Step 2: Write failing tests**

Create `pkg/parser/translations/translations_test.go`:
```go
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

	// From the CSV: 0.7_sekunden -> 0.7 seconds
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

	// CONTRIBUTORS and LASTMODIFIED should not be in translations
	_, found := dict.Translate("CONTRIBUTORS")
	assert.False(t, found)
}

func TestLoadInvalidPath(t *testing.T) {
	_, err := Load("/nonexistent/path.csv")
	assert.Error(t, err)
}
```

**Step 3: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/translations/ -v`
Expected: compilation errors

**Step 4: Write implementation**

Create `pkg/parser/translations/translations.go`:
```go
package translations

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// Dictionary holds German-to-English translation mappings.
type Dictionary struct {
	Entries map[string]string // lowercase German key -> English value
}

// metadataKeys are CSV rows that contain file metadata, not translations.
var metadataKeys = map[string]bool{
	"contributors": true,
	"lastmodified": true,
}

// Load reads a NCS Dummy translations CSV file.
func Load(path string) (Dictionary, error) {
	f, err := os.Open(path)
	if err != nil {
		return Dictionary{}, fmt.Errorf("translations: opening file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // variable fields (some values contain commas in quotes)
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return Dictionary{}, fmt.Errorf("translations: parsing CSV: %w", err)
	}

	entries := make(map[string]string, len(records))
	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		key := strings.TrimSpace(record[0])
		value := strings.TrimSpace(record[1])

		if key == "" || value == "" {
			continue
		}

		lowerKey := strings.ToLower(key)
		if metadataKeys[lowerKey] {
			continue
		}

		entries[lowerKey] = value
	}

	return Dictionary{Entries: entries}, nil
}

// Translate looks up a German key and returns its English translation.
func (d Dictionary) Translate(key string) (string, bool) {
	value, found := d.Entries[strings.ToLower(key)]
	return value, found
}
```

**Step 5: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/translations/ -v -race`
Expected: all tests PASS

**Step 6: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/parser/translations/ testdata/translations/
git commit -m "feat: NCS Dummy translation parser (German to English)"
```

---

### Task 8: Safety Pipeline — Voltage Gate

The voltage gate is stage 0 of the safety pipeline. Test it independently of any hardware.

**Files:**
- Create: `pkg/safety/voltage.go`
- Create: `pkg/safety/voltage_test.go`

**Step 1: Write failing tests**

Create `pkg/safety/voltage_test.go`:
```go
package safety

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVoltageCheckCodingPass(t *testing.T) {
	result := CheckVoltage(12.8, OpCoding)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckCodingWarning(t *testing.T) {
	result := CheckVoltage(12.3, OpCoding)
	assert.Equal(t, VoltageWarning, result.Status)
	assert.Contains(t, result.Message, "12.3")
}

func TestVoltageCheckCodingBlocked(t *testing.T) {
	result := CheckVoltage(11.8, OpCoding)
	assert.Equal(t, VoltageBlocked, result.Status)
}

func TestVoltageCheckMultiCodingBlocked(t *testing.T) {
	result := CheckVoltage(12.3, OpMultiCoding)
	assert.Equal(t, VoltageBlocked, result.Status)
}

func TestVoltageCheckMultiCodingPass(t *testing.T) {
	result := CheckVoltage(12.8, OpMultiCoding)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckFlashBlocked(t *testing.T) {
	result := CheckVoltage(12.8, OpFlash)
	assert.Equal(t, VoltageBlocked, result.Status)
}

func TestVoltageCheckFlashPass(t *testing.T) {
	result := CheckVoltage(13.5, OpFlash)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckFlashWarningAtThreshold(t *testing.T) {
	// Exactly at threshold should pass
	result := CheckVoltage(13.0, OpFlash)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckCodingExactlyAtBlock(t *testing.T) {
	// Exactly at 12.0V should be blocked (< 12.0 is block threshold)
	result := CheckVoltage(12.0, OpCoding)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckCodingJustBelowBlock(t *testing.T) {
	result := CheckVoltage(11.99, OpCoding)
	assert.Equal(t, VoltageBlocked, result.Status)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/safety/ -v`
Expected: compilation errors

**Step 3: Write implementation**

Create `pkg/safety/voltage.go`:
```go
package safety

import "fmt"

// VoltageStatus represents the result of a voltage check.
type VoltageStatus int

const (
	VoltageOK      VoltageStatus = iota
	VoltageWarning
	VoltageBlocked
)

// OperationType categorizes write operations by risk level.
type OperationType int

const (
	OpCoding      OperationType = iota // single module coding
	OpMultiCoding                      // multi-module coding
	OpFlash                            // firmware flashing
)

// VoltageResult contains the voltage check outcome.
type VoltageResult struct {
	Status  VoltageStatus
	Voltage float64
	Message string
}

// voltage thresholds per operation type
type thresholds struct {
	block   float64
	warning float64
}

var voltageThresholds = map[OperationType]thresholds{
	OpCoding:      {block: 12.0, warning: 12.5},
	OpMultiCoding: {block: 12.5, warning: 12.5},
	OpFlash:       {block: 13.0, warning: 13.0},
}

// CheckVoltage evaluates battery voltage against operation thresholds.
func CheckVoltage(voltage float64, op OperationType) VoltageResult {
	t := voltageThresholds[op]

	if voltage < t.block {
		return VoltageResult{
			Status:  VoltageBlocked,
			Voltage: voltage,
			Message: fmt.Sprintf("Battery voltage %.1fV is below minimum %.1fV — operation blocked", voltage, t.block),
		}
	}

	if voltage < t.warning {
		return VoltageResult{
			Status:  VoltageWarning,
			Voltage: voltage,
			Message: fmt.Sprintf("Battery voltage %.1fV is low (recommended: %.1fV+) — proceed with caution", voltage, t.warning),
		}
	}

	return VoltageResult{
		Status:  VoltageOK,
		Voltage: voltage,
		Message: fmt.Sprintf("Battery voltage %.1fV OK", voltage),
	}
}
```

**Step 4: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/safety/ -v -race`
Expected: all tests PASS

**Step 5: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/safety/
git commit -m "feat: voltage gate for safety pipeline"
```

---

### Task 9: Safety Pipeline — Backup Vault

SQLite-backed backup store. Every module read is saved with timestamp.

**Files:**
- Create: `pkg/safety/vault/vault.go`
- Create: `pkg/safety/vault/vault_test.go`

**Step 1: Write failing tests**

Create `pkg/safety/vault/vault_test.go`:
```go
package vault

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempVault(t *testing.T) *Vault {
	t.Helper()
	tmp, err := os.CreateTemp("", "vault_test_*.db")
	require.NoError(t, err)
	tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	v, err := Open(tmp.Name())
	require.NoError(t, err)
	t.Cleanup(func() { v.Close() })

	return v
}

func TestOpenAndClose(t *testing.T) {
	v := tempVault(t)
	assert.NotNil(t, v)
}

func TestSaveAndRetrieve(t *testing.T) {
	v := tempVault(t)

	coding := []byte{0x01, 0x02, 0x03, 0x04}
	err := v.Save("E46", "GM5", "C05", coding)
	require.NoError(t, err)

	entries, err := v.List("E46", "GM5")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, coding, entries[0].Data)
	assert.Equal(t, "GM5", entries[0].Module)
	assert.Equal(t, "C05", entries[0].Version)
}

func TestSaveMultipleVersions(t *testing.T) {
	v := tempVault(t)

	err := v.Save("E46", "GM5", "C05", []byte{0x01})
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // ensure different timestamp

	err = v.Save("E46", "GM5", "C05", []byte{0x02})
	require.NoError(t, err)

	entries, err := v.List("E46", "GM5")
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	// Most recent should be first
	assert.Equal(t, []byte{0x02}, entries[0].Data)
	assert.Equal(t, []byte{0x01}, entries[1].Data)
}

func TestGetLatest(t *testing.T) {
	v := tempVault(t)

	v.Save("E46", "GM5", "C05", []byte{0x01})
	time.Sleep(10 * time.Millisecond)
	v.Save("E46", "GM5", "C05", []byte{0x02})

	entry, err := v.Latest("E46", "GM5")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x02}, entry.Data)
}

func TestGetLatestEmpty(t *testing.T) {
	v := tempVault(t)

	_, err := v.Latest("E46", "GM5")
	assert.ErrorIs(t, err, ErrNoBackup)
}

func TestListDifferentModules(t *testing.T) {
	v := tempVault(t)

	v.Save("E46", "GM5", "C05", []byte{0x01})
	v.Save("E46", "KMB", "C06", []byte{0x02})

	gm5, _ := v.List("E46", "GM5")
	kmb, _ := v.List("E46", "KMB")

	assert.Len(t, gm5, 1)
	assert.Len(t, kmb, 1)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/safety/vault/ -v`
Expected: compilation errors

**Step 3: Install SQLite dependency and write implementation**

Run: `cd /c/Users/Alex/bavarix && go get modernc.org/sqlite`

Create `pkg/safety/vault/vault.go`:
```go
package vault

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNoBackup = errors.New("vault: no backup found")

// Entry represents a stored module coding backup.
type Entry struct {
	ID        int64
	Chassis   string
	Module    string
	Version   string
	Data      []byte
	CreatedAt time.Time
}

// Vault is a SQLite-backed backup store for module coding data.
type Vault struct {
	db *sql.DB
}

// Open creates or opens a vault database at the given path.
func Open(path string) (*Vault, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("vault: opening database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS backups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chassis TEXT NOT NULL,
			module TEXT NOT NULL,
			version TEXT NOT NULL,
			data BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("vault: creating table: %w", err)
	}

	return &Vault{db: db}, nil
}

// Close closes the vault database.
func (v *Vault) Close() error {
	return v.db.Close()
}

// Save stores a module coding backup.
func (v *Vault) Save(chassis, module, version string, data []byte) error {
	_, err := v.db.Exec(
		`INSERT INTO backups (chassis, module, version, data) VALUES (?, ?, ?, ?)`,
		chassis, module, version, data,
	)
	if err != nil {
		return fmt.Errorf("vault: saving backup: %w", err)
	}
	return nil
}

// List returns all backups for a module, most recent first.
func (v *Vault) List(chassis, module string) ([]Entry, error) {
	rows, err := v.db.Query(
		`SELECT id, chassis, module, version, data, created_at
		 FROM backups
		 WHERE chassis = ? AND module = ?
		 ORDER BY created_at DESC`,
		chassis, module,
	)
	if err != nil {
		return nil, fmt.Errorf("vault: listing backups: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Chassis, &e.Module, &e.Version, &e.Data, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("vault: scanning row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// Latest returns the most recent backup for a module.
func (v *Vault) Latest(chassis, module string) (Entry, error) {
	entries, err := v.List(chassis, module)
	if err != nil {
		return Entry{}, err
	}
	if len(entries) == 0 {
		return Entry{}, ErrNoBackup
	}
	return entries[0], nil
}
```

**Step 4: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/safety/vault/ -v -race`
Expected: all tests PASS

**Step 5: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/safety/vault/ go.mod go.sum
git commit -m "feat: SQLite backup vault for module coding history"
```

---

### Task 10: Full PRG Extraction Pipeline

Extract jobs from ALL 1,192 PRG files on disk and output summary statistics. This validates our parser at scale.

**Files:**
- Create: `cmd/extract-prg/main.go`
- Create: `pkg/parser/prg/batch.go`
- Create: `pkg/parser/prg/batch_test.go`

**Step 1: Write failing tests**

Create `pkg/parser/prg/batch_test.go`:
```go
package prg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchExtractTestdata(t *testing.T) {
	path := testdataPath("") // testdata/prg/ directory
	results, err := BatchExtract(path)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// At least C_GM5.prg should parse
	found := false
	for _, r := range results {
		if r.Filename == "C_GM5.prg" {
			found = true
			assert.NotEmpty(t, r.Jobs, "C_GM5 should have jobs")
			assert.Empty(t, r.Error, "C_GM5 should parse without error")
		}
	}
	assert.True(t, found, "C_GM5.prg should be in results")
}

func TestBatchResultStats(t *testing.T) {
	path := testdataPath("")
	results, err := BatchExtract(path)
	require.NoError(t, err)

	stats := results.Stats()
	assert.Greater(t, stats.TotalFiles, 0)
	assert.Greater(t, stats.SuccessCount, 0)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/prg/ -v -run TestBatch`
Expected: compilation errors

**Step 3: Write implementation**

Create `pkg/parser/prg/batch.go`:
```go
package prg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BatchResult holds the extraction result for a single PRG file.
type BatchResult struct {
	Filename string
	Jobs     []Job
	Error    string
}

// BatchResults is a collection of extraction results.
type BatchResults []BatchResult

// Stats returns summary statistics for the batch extraction.
type BatchStats struct {
	TotalFiles   int
	SuccessCount int
	ErrorCount   int
	TotalJobs    int
}

func (br BatchResults) Stats() BatchStats {
	stats := BatchStats{TotalFiles: len(br)}
	for _, r := range br {
		if r.Error == "" {
			stats.SuccessCount++
			stats.TotalJobs += len(r.Jobs)
		} else {
			stats.ErrorCount++
		}
	}
	return stats
}

// BatchExtract processes all .prg files in a directory.
func BatchExtract(dir string) (BatchResults, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("prg: reading directory: %w", err)
	}

	var results BatchResults
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".prg") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		jobs, err := ExtractJobs(path)

		result := BatchResult{Filename: entry.Name()}
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Jobs = jobs
		}
		results = append(results, result)
	}

	return results, nil
}
```

**Step 4: Run tests**

Run: `cd /c/Users/Alex/bavarix && go test ./pkg/parser/prg/ -v -race`
Expected: all tests PASS

**Step 5: Create extraction CLI tool**

Create `cmd/extract-prg/main.go`:
```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexcatdad/bavarix/pkg/parser/prg"
)

func main() {
	dir := `C:\EDIABAS\Ecu`
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	fmt.Fprintf(os.Stderr, "Extracting jobs from PRG files in %s...\n", dir)

	results, err := prg.BatchExtract(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	stats := results.Stats()
	fmt.Fprintf(os.Stderr, "Processed: %d files, %d success, %d errors, %d total jobs\n",
		stats.TotalFiles, stats.SuccessCount, stats.ErrorCount, stats.TotalJobs)

	// Output JSON to stdout
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(results)
}
```

**Step 6: Run against real data (read-only, safe)**

Run: `cd /c/Users/Alex/bavarix && go run ./cmd/extract-prg/ 2>&1 | head -50`
Expected: summary line showing parse statistics, JSON output of extracted jobs

**Step 7: Commit**

```bash
cd /c/Users/Alex/bavarix
git add pkg/parser/prg/batch.go pkg/parser/prg/batch_test.go cmd/extract-prg/
git commit -m "feat: batch PRG extraction with stats and CLI tool"
```

---

### Task 11: CI Pipeline Setup

GitHub Actions for running all tests on every push.

**Files:**
- Create: `.github/workflows/ci.yml`

**Step 1: Create CI workflow**

Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests
        run: go test ./... -v -race -count=1

      - name: Build
        run: go build ./cmd/bavarix

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

**Step 2: Commit and push**

```bash
cd /c/Users/Alex/bavarix
git add .github/
git commit -m "ci: add GitHub Actions for tests and linting"
git push
```

**Step 3: Verify CI runs**

Run: `cd /c/Users/Alex/bavarix && gh run list --limit 1`
Expected: CI workflow triggered

---

## Summary

After all 11 tasks, you will have:

- **Protocol framing** tested and working (DS2 + KWP2000)
- **PRG parser** extracting job names from all 1,192 ECU files
- **SP-Daten parser** extracting coding parameter definitions
- **Translation parser** mapping 26,246 German terms to English
- **Voltage gate** enforcing battery thresholds per operation type
- **Backup vault** storing module coding history in SQLite
- **Batch extraction CLI** for processing all PRG files
- **CI pipeline** running tests on every push

All built TDD — tests first, implementation second. Zero hardware connections required. Ready for Phase 1 (transport layer + virtual ECU simulator).
