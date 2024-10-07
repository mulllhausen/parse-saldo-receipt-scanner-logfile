package main

import (
	"os"
	saldoProcessor "parse-saldo/grunt"
	"path/filepath"
	"strings"
)

func main() {
	baseLogfileName := "logfile.log" // default
	if len(os.Args) > 1 {
		baseLogfileName = os.Args[1]
	}
	fileName := strings.TrimSuffix(baseLogfileName, filepath.Ext(baseLogfileName))
	args := saldoProcessor.ConvertLogsToCSVArgs{
		Logfile:          fileName + ".log",
		OutputToFile:     true,
		CSVFile:          fileName + ".csv",
		RemoveDuplicates: true,
		SortByDate:       true,
	}
	saldoProcessor.ConvertLogsToCSV(args)
}
