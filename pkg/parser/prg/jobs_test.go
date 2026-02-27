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

	jobNames := make(map[string]bool)
	for _, j := range jobs {
		jobNames[j.Name] = true
	}

	// GM5 should have common diagnostic jobs
	assert.True(t, len(jobNames) > 1, "should extract multiple job names")

	// Verify specific known job names
	assert.True(t, jobNames["INFO"], "should contain INFO job")
	assert.True(t, jobNames["IDENT"], "should contain IDENT job")
	assert.True(t, jobNames["INITIALISIERUNG"], "should contain INITIALISIERUNG job")
}

func TestExtractJobsFromLSZ(t *testing.T) {
	path := testdataPath("LSZ.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	require.NotEmpty(t, jobs, "LSZ PRG should contain jobs")

	jobNames := make(map[string]bool)
	for _, j := range jobs {
		jobNames[j.Name] = true
	}

	// LSZ has many jobs including some specific ones
	assert.True(t, jobNames["FS_LESEN"], "should contain FS_LESEN job")
	assert.True(t, jobNames["SPEICHER_LESEN"], "should contain SPEICHER_LESEN job")
	assert.True(t, jobNames["POWER_DOWN"], "should contain POWER_DOWN job")
}

func TestExtractJobsFromKMB46(t *testing.T) {
	path := testdataPath("C_KMB46.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	require.NotEmpty(t, jobs, "KMB46 PRG should contain jobs")

	jobNames := make(map[string]bool)
	for _, j := range jobs {
		jobNames[j.Name] = true
	}

	assert.True(t, jobNames["C_FG_LESEN"], "should contain C_FG_LESEN job")
	assert.True(t, jobNames["SOFTWARE_RESET"], "should contain SOFTWARE_RESET job")
}

func TestExtractJobsInvalidFile(t *testing.T) {
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

func TestExtractJobsCount(t *testing.T) {
	path := testdataPath("C_GM5.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	// GM5 should have exactly 11 jobs based on the embedded count
	assert.Equal(t, 11, len(jobs), "GM5 should have 11 jobs")
}

func TestExtractJobsLSZCount(t *testing.T) {
	path := testdataPath("LSZ.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	// LSZ should have exactly 30 jobs
	assert.Equal(t, 30, len(jobs), "LSZ should have 30 jobs")
}

func TestExtractJobsKMB46Count(t *testing.T) {
	path := testdataPath("C_KMB46.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	jobs, err := ExtractJobs(path)
	require.NoError(t, err)
	// KMB46 should have exactly 22 jobs
	assert.Equal(t, 22, len(jobs), "KMB46 should have 22 jobs")
}

func TestExtractJobsHaveCodeAddresses(t *testing.T) {
	path := testdataPath("C_GM5.prg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test PRG file not available")
	}

	jobs, err := ExtractJobs(path)
	require.NoError(t, err)

	// Each job (except possibly the first) should have a non-zero address
	for _, j := range jobs[1:] {
		assert.NotZero(t, j.Address, "job %s should have a non-zero code address", j.Name)
	}
}
