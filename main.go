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
	saldoProcessor.ConvertLogsToCSV(fileName+".log", fileName+".csv")
}
