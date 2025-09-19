package tools

import (
	"encoding/csv"
	"fmt"
	"os"
)

func GetLastNRecords(filePath string, n int) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open csv file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read csv file: %w", err)
	}

	totalRecords := len(records)
	if totalRecords == 0 {
		return [][]string{}, nil
	}

	startIndex := totalRecords - n
	if startIndex < 0 {
		startIndex = 0
	}

	return records[startIndex:], nil
}
