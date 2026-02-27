package translations

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

type Dictionary struct {
	Entries map[string]string
}

var metadataKeys = map[string]bool{
	"contributors": true,
	"lastmodified": true,
}

func Load(path string) (Dictionary, error) {
	f, err := os.Open(path)
	if err != nil {
		return Dictionary{}, fmt.Errorf("translations: opening file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
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

func (d Dictionary) Translate(key string) (string, bool) {
	value, found := d.Entries[strings.ToLower(key)]
	return value, found
}
