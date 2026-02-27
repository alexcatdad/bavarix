package prg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchExtractTestdata(t *testing.T) {
	path := testdataPath("")
	results, err := BatchExtract(path)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

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

func TestBatchExtractAllFilesProcessed(t *testing.T) {
	path := testdataPath("")
	results, err := BatchExtract(path)
	require.NoError(t, err)
	// We have C_GM5.prg, LSZ.prg, and C_KMB46.prg in testdata
	assert.Len(t, results, 3, "should process all test PRG files")

	filenames := make(map[string]bool)
	for _, r := range results {
		filenames[r.Filename] = true
	}
	assert.True(t, filenames["C_GM5.prg"], "should include C_GM5.prg")
	assert.True(t, filenames["LSZ.prg"], "should include LSZ.prg")
	assert.True(t, filenames["C_KMB46.prg"], "should include C_KMB46.prg")
}

func TestBatchResultStats(t *testing.T) {
	path := testdataPath("")
	results, err := BatchExtract(path)
	require.NoError(t, err)

	stats := results.Stats()
	assert.Equal(t, 3, stats.TotalFiles, "should have 3 total files")
	assert.Equal(t, 3, stats.SuccessCount, "all files should parse successfully")
	assert.Equal(t, 0, stats.ErrorCount, "should have no errors")
	assert.Greater(t, stats.TotalJobs, 0, "should have extracted some jobs")
}

func TestBatchExtractEmptyDir(t *testing.T) {
	tmp := t.TempDir()
	results, err := BatchExtract(tmp)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestBatchExtractNonexistentDir(t *testing.T) {
	_, err := BatchExtract("/nonexistent/directory/path")
	assert.Error(t, err)
}

func TestBatchResultStatsWithErrors(t *testing.T) {
	results := BatchResults{
		{Filename: "good.prg", Jobs: []Job{{Name: "JOB1"}, {Name: "JOB2"}}},
		{Filename: "bad.prg", Error: "some error"},
		{Filename: "good2.prg", Jobs: []Job{{Name: "JOB3"}}},
	}
	stats := results.Stats()
	assert.Equal(t, 3, stats.TotalFiles)
	assert.Equal(t, 2, stats.SuccessCount)
	assert.Equal(t, 1, stats.ErrorCount)
	assert.Equal(t, 3, stats.TotalJobs)
}

func TestBatchExtractSkipsNonPRGFiles(t *testing.T) {
	tmp := t.TempDir()
	// Create a non-PRG file
	require.NoError(t, writeFile(tmp+"/readme.txt", []byte("hello")))
	// Create a valid PRG file by copying header
	header := make([]byte, 0xA0)
	copy(header[:16], MagicHeader)
	require.NoError(t, writeFile(tmp+"/test.prg", header))

	results, err := BatchExtract(tmp)
	require.NoError(t, err)
	assert.Len(t, results, 1, "should only process .prg files")
	assert.Equal(t, "test.prg", results[0].Filename)
}

func writeFile(path string, data []byte) error {
	return writeFileHelper(path, data)
}
