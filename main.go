package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	baseLogfileName := "logfile.log" // default
	if len(os.Args) > 1 {
		baseLogfileName = os.Args[1]
		fmt.Println("Using log file: ", baseLogfileName)
	}
	fileName := strings.TrimSuffix(baseLogfileName, filepath.Ext(baseLogfileName))
	ConvertLogsToCsv(fileName+".log", fileName+".csv")
}
