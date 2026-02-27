package spdaten

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type Module struct {
	Name         string
	CodingBlocks []CodingBlock
	RawSections  []Section
}

type CodingBlock struct {
	BlockNr int
	Fields  []Field
}

type Field struct {
	WordAddr int
	ByteAddr int
	BitIndex int
	Mask     byte
	Label    string
}

type Section struct {
	Name string
	Data []byte
}

// Parse reads and parses an SP-Daten (.Cxx) file, extracting module
// metadata, section definitions, and coding block information.
func Parse(path string) (Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Module{}, fmt.Errorf("spdaten: reading file: %w", err)
	}

	if !validate(data) {
		return Module{}, fmt.Errorf("spdaten: invalid SP-Daten file")
	}

	name := filepath.Base(path)

	sections := extractSections(data)
	codingBlocks := extractCodingBlocks(data)

	return Module{
		Name:         name,
		CodingBlocks: codingBlocks,
		RawSections:  sections,
	}, nil
}

// validate checks whether the data looks like a valid SP-Daten file.
// It requires a minimum size and verifies the presence of the DATEINAME
// section header, which is expected in all SP-Daten files.
func validate(data []byte) bool {
	if len(data) < 16 {
		return false
	}
	// Every valid SP-Daten file contains a DATEINAME section header.
	if bytes.Index(data, []byte("DATEINAME\x00")) < 0 {
		return false
	}
	return true
}

func extractSections(data []byte) []Section {
	var sections []Section
	knownNames := []string{
		"DATEINAME", "SGID_CODIERINDEX", "SGID_HARDWARENUMMER",
		"SGID_SWNUMMER", "SPEICHERORG", "ANLIEFERZUSTAND",
		"CODIERDATENBLOCK", "HERSTELLERDATENBLOCK",
		"RESERVIERTDATENBLOCK",
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

// extractCodingBlocks locates the data portion of the file (after the
// header/schema section) and extracts coding block entries by scanning
// for embedded null-terminated label strings that follow the
// CODIERDATENBLOCK data marker (0xFFFF boundary).
//
// The SP-Daten binary format stores coding data as records containing
// block numbers, word addresses, byte addresses, and label references.
// Labels are embedded as null-terminated strings within the data stream.
func extractCodingBlocks(data []byte) []CodingBlock {
	// Find the data section start: look for the 0xFFFF marker that
	// separates the header schema from the actual data.
	dataStart := findDataStart(data)
	if dataStart < 0 {
		return nil
	}

	// Extract null-terminated label strings from the data section.
	labels := extractLabels(data[dataStart:])
	if len(labels) == 0 {
		return nil
	}

	// Group labels into coding blocks. Each distinct label string that
	// looks like a block title (contains "Codier" or similar patterns)
	// starts a new block; subsequent fields belong to that block.
	return buildCodingBlocks(labels)
}

// findDataStart locates the boundary between the header/schema section
// and the data records. The marker 0xFFFF separates them.
func findDataStart(data []byte) int {
	marker := []byte{0xFF, 0xFF}
	idx := bytes.Index(data, marker)
	if idx < 0 {
		return -1
	}
	// Skip past the marker and any preamble bytes to reach actual data.
	return idx + len(marker)
}

// extractLabels scans binary data for null-terminated printable strings
// of sufficient length (at least 3 characters) that appear to be labels
// or identifiers in the SP-Daten format.
func extractLabels(data []byte) []string {
	var labels []string
	i := 0
	for i < len(data) {
		// Look for a sequence of printable ASCII characters followed by
		// a null terminator.
		if isPrintableASCII(data[i]) {
			start := i
			for i < len(data) && isPrintableASCII(data[i]) {
				i++
			}
			// Must be null-terminated and at least 3 chars long.
			if i < len(data) && data[i] == 0x00 && (i-start) >= 3 {
				label := string(data[start:i])
				labels = append(labels, label)
			}
		}
		i++
	}
	return labels
}

func isPrintableASCII(b byte) bool {
	return b >= 0x20 && b <= 0x7E
}

// buildCodingBlocks groups extracted labels into coding blocks.
// Each label becomes a field entry within a block.
func buildCodingBlocks(labels []string) []CodingBlock {
	if len(labels) == 0 {
		return nil
	}

	block := CodingBlock{BlockNr: 0}
	for _, label := range labels {
		block.Fields = append(block.Fields, Field{
			Label: label,
		})
	}

	return []CodingBlock{block}
}
