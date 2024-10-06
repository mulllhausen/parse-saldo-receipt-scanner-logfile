# parse-saldo-receipt-scanner-logfile
Parse the logfile of the Saldo Receipt Scanner Android app

The Saldo Receipt Scanner app is the best I have found for parsing
receipt text, including the lines. However their export feature does
not include the receipt lines - only the header info like total, date,
merchant, etc.

This repo is a workaround for getting the receipt lines.

To get the logfile (last time I checked):
- go to the home page
- press the `... more` menu (bottom right)
- press `Help Center`
- a zip file appears. email it to yourself.
- download the zip file
- extract it
- inside there is a `.log` file with all your data

The `.log` file is unstructured data. This script parses it and extracts
the data into a CSV file. [Install golang](https://go.dev/dl/) then run
it like this:

    go run parse_log.go

if you want debugging:

    go install github.com/go-delve/delve/cmd/dlv@latest