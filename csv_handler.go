package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// ReadCSV reads a CSV file and returns the headers and records
func ReadCSV(path string) ([]string, [][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read CSV: %v", err)
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("csv file is empty")
	}

	headers := records[0]
	data := records[1:]

	return headers, data, nil
}
