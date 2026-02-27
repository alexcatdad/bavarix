package prg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BatchResult holds the extraction result for a single PRG file.
type BatchResult struct {
	Filename string `json:"filename"`
	Jobs     []Job  `json:"jobs,omitempty"`
	Error    string `json:"error,omitempty"`
}

// BatchResults is a collection of batch extraction results.
type BatchResults []BatchResult

// BatchStats holds aggregate statistics for a batch extraction run.
type BatchStats struct {
	TotalFiles   int `json:"total_files"`
	SuccessCount int `json:"success_count"`
	ErrorCount   int `json:"error_count"`
	TotalJobs    int `json:"total_jobs"`
}

// Stats computes aggregate statistics over the batch results.
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

// BatchExtract reads all .prg files in the given directory and extracts
// jobs from each one. Non-.prg files and subdirectories are skipped.
// Individual file errors are captured in the result rather than aborting
// the entire batch.
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
