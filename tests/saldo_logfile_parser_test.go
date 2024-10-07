package tests

import (
	"fmt"
	"os"
	saldoProcessor "parse-saldo/grunt"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllFiles(t *testing.T) {
	// arrange
	testDirectory := "."

	singleTestNumber := 0 // 0 = test all files
	singleTestFilePrefix := fmt.Sprintf("test%03d", singleTestNumber)

	err := filepath.Walk(testDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != ".log" {
			return nil
		}
		filename := filepath.Base(path)
		baseFilename := strings.TrimSuffix(filename, ".log")

		if singleTestNumber != 0 &&
			!strings.HasPrefix(filename, singleTestFilePrefix) {
			return nil
		}

		expectedCSVFilename := filepath.Join(testDirectory, baseFilename+".csv")
		expectedCSVDataBytes, err := os.ReadFile(expectedCSVFilename)
		if err != nil {
			t.Fatalf("Error reading expected CSV file %s: %v", expectedCSVFilename, err)
		}
		expectedCSVData := strings.ReplaceAll(string(expectedCSVDataBytes), "\r", "")

		// act
		actualCSVData := saldoProcessor.ConvertLogsToCSV(path, "")

		// assert
		if strings.TrimRight(actualCSVData, "\n") != strings.TrimRight(expectedCSVData, "\n") {
			t.Errorf("CSV content mismatch for %v: expected vs generated", baseFilename)
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
