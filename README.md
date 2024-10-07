# parse-saldo-receipt-scanner-logfile
Parse the logfile of the Saldo Receipt Scanner Android app

The Saldo Receipt Scanner app is the best I have found for parsing
receipt text, including the lines. However their CSV export feature does
not include the receipt lines - only the header info like total, date,
merchant, etc. The receipt lines are added to the logfile when they are
processed by the Saldo App, so this project parses the logfile to output 
a CSV containing the receipt lines.

To get the logfile out of the app is not obvious:
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

    go run . processing/test-logfile.log

it will output into the same `processing` dir.

build:

    go build -o parse-saldo-logfile.exe

run tests:

    go test ./tests/

if you want debugging in vscode:

    go install github.com/go-delve/delve/cmd/dlv@latest
