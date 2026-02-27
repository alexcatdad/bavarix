package prg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractJobsCGM5(t *testing.T) {
	path := testdataPath("C_GM5.prg")
	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	require.NotEmpty(t, jobs, "C_GM5.prg should contain jobs")

	// We know from investigation that C_GM5 has these jobs:
	expectedJobs := []string{
		"INITIALISIERUNG",
		"IDENT",
		"DIAGNOSE_ENDE",
		"C_FG_LESEN",
		"C_FG_AUFTRAG",
		"C_C_LESEN",
		"C_C_AUFTRAG",
		"STATUS_DIGITAL",
		"STATUS_FUNKSCHLUESSEL",
		"STATUS_KEY_MEMORY",
	}

	jobNames := make([]string, len(jobs))
	for i, j := range jobs {
		jobNames[i] = j.Name
	}

	assert.Equal(t, expectedJobs, jobNames, "job names should match expected list")
	assert.Len(t, jobs, 10, "C_GM5 should have exactly 10 jobs")
}

func TestExtractJobsLSZ(t *testing.T) {
	path := testdataPath("LSZ.prg")
	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	require.NotEmpty(t, jobs, "LSZ.prg should contain jobs")

	// Verify common expected jobs exist
	jobNames := make(map[string]bool)
	for _, j := range jobs {
		jobNames[j.Name] = true
	}
	assert.True(t, jobNames["IDENT"], "LSZ should have IDENT job")
}

func TestExtractJobsInvalidFile(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/invalid.prg"
	err := os.WriteFile(path, []byte("not a prg file"), 0644)
	require.NoError(t, err)

	_, err = ExtractJobs(path)
	assert.ErrorIs(t, err, ErrInvalidMagic)
}

func TestExtractJobsNonexistentFile(t *testing.T) {
	_, err := ExtractJobs("/nonexistent/path/file.prg")
	assert.Error(t, err)
}

func TestExtractJobsCodeOffsets(t *testing.T) {
	path := testdataPath("C_GM5.prg")
	jobs, err := ExtractJobs(path)
	require.NoError(t, err)

	for _, j := range jobs {
		assert.NotEmpty(t, j.Name, "job name should not be empty")
		// Code offsets should be reasonable values within the file
		assert.Greater(t, j.CodeOffset, uint32(0), "code offset for %s should be > 0", j.Name)
	}
}
