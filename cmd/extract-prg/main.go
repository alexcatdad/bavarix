package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexcatdad/bavarix/pkg/parser/prg"
)

func main() {
	dir := `C:\Users\Alex\bavarix\data\ediabas\ecu`
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

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
