package tests

import (
	"os"
	saldoProcessor "parse-saldo/grunt"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllHappyPathFiles(t *testing.T) {
	// arrange
	testDirectory := "happy-path"
	err := filepath.Walk(testDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != ".log" {
			return nil
		}
		baseFilename := strings.TrimSuffix(filepath.Base(path), ".log")
		expectedCSVFilename := filepath.Join(testDirectory, baseFilename+".csv")
		expectedCsvDataBytes, err := os.ReadFile(expectedCSVFilename)
		if err != nil {
			t.Fatalf("Error reading expected CSV file %s: %v", expectedCSVFilename, err)
		}
		expectedCsvData := strings.ReplaceAll(string(expectedCsvDataBytes), "\r", "")

		// act
		actualCSVData := saldoProcessor.ConvertLogsToCSV(path, "")

		// assert
		if strings.TrimRight(actualCSVData, "\n") != strings.TrimRight(expectedCsvData, "\n") {
			t.Errorf("CSV content mismatch for %v: expected vs generated", baseFilename)
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
