package saldoProcessor

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Receipt struct {
	LineNumber   int
	Date         string
	Total        string
	Currency     string
	Merchant     string
	Category     string
	Description  string
	IsReceipt    bool
	IsFirstEvent bool
	Title        string
	Name         string
	IsReconciled bool
	ReceiptLines []ReceiptLine
}

type ReceiptLine struct {
	Name         string
	Quantity     string
	PricePerUnit string
	TotalPrice   string
}

// only writes to CSV file when csvFile is supplied
func ConvertLogsToCSV(logfile string, csvFile string) string {
	receipts, err := processEntireLogFile(logfile)
	if err != nil {
		fmt.Printf("Error reading log file %s: %v\n", logfile, err)
		return ""
	}
	csv := toCSV(receipts)
	if csvFile != "" {
		writeToFile(csv, csvFile)
	}
	return csv
}

func processEntireLogFile(filename string) ([]Receipt, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var receipts []Receipt
	scanner := bufio.NewScanner(file)
	var currentRecord strings.Builder

	// match a date at the start of a line
	dateRegex := regexp.MustCompile(`^\d{2}-\d{2}-\d{4}`)

	lineNumber := 1      // keep track of the logfile line
	datedLineNumber := 1 // use in the output CSV
	for scanner.Scan() {
		line := scanner.Text()

		if dateRegex.MatchString(line) {
			// process old record before starting a new one
			if currentRecord.String() != "" {
				receipt, err := parseReceipt(currentRecord.String())
				if err != nil {
					fmt.Printf("Error parsing line %d: %v\n", lineNumber, err)
				}
				receipt.LineNumber = datedLineNumber
				if receipt.IsReceipt {
					receipts = append(receipts, receipt)
				}
			}
			// start a new record
			currentRecord.Reset()
			currentRecord.WriteString(line)
			datedLineNumber = lineNumber
		} else {
			currentRecord.WriteString(" " + line)
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// process the final record
	if currentRecord.String() != "" {
		receipt, err := parseReceipt(currentRecord.String())
		if err != nil {
			fmt.Printf("Error parsing line %d: %v\n", lineNumber, err)
		}
		receipt.LineNumber = datedLineNumber
		if receipt.IsReceipt {
			receipts = append(receipts, receipt)
		}
	}
	return receipts, nil
}

func parseReceipt(logLine string) (Receipt, error) {
	// in the logfile the props line looks like this, without the newlines:
	// props: {
	//      date=1711627200000,
	//      total=29.9,
	//      currency=AUD,
	//      merchant=Coles Supermarkets Australia Pty Ltd coles,
	//      category=,
	//      description=,
	//      receipt=true,
	//      place=onboarding,
	//      items=
	//          1_Item
	//              name: WELLNESS ROAD LINSEE 500GRAM 3 @ $3.30 EACH,
	//              quantity: 0.0,
	//              pricePerUnit: ,
	//              totalPrice: 990
	//          2_Item
	//              name: % TULIPS 1EACH,
	//              quantity: 0.0,
	//              pricePerUnit: ,
	//              totalPrice: 2000,
	//      first_event=true
	// }
	//
	// note:
	// - we cannot be sure that "first_event" will always come last. assume keys can be
	// in any order
	// - we cannot immediately split by "," because we want the whole of "items" to be 1
	// keyvalue pair since this is a nested list

	logLine = parseCleanProps(logLine)
	if logLine == "" {
		return Receipt{}, nil
	}

	receipt := Receipt{} // init

	// note: this is necessary because golang does not support regex lookaheads
	mismatchingPairs := strings.Split(logLine, "=")
	numMismatchingPairs := len(mismatchingPairs)
	for i, mismatchingPair := range mismatchingPairs {
		if i == 0 {
			continue
		}
		isLastItem := i == (numMismatchingPairs - 1)
		previousWords := strings.Split(mismatchingPairs[i-1], ",")
		previousLastWord := previousWords[len(previousWords)-1]
		currentWords := strings.Split(mismatchingPair, ",")

		// chop off the last current word
		if !isLastItem {
			currentWords = currentWords[:len(currentWords)-1]
		}

		key := strings.TrimSpace(previousLastWord)
		value := strings.Join(currentWords, ",")
		value = strings.TrimRight(value, ",\n\r ")
		if value == "in progress" {
			return Receipt{}, nil
		}

		switch key {
		case "date":
			receipt.Date = parseUnixtime(value)
		case "total":
			receipt.Total = strings.ReplaceAll(value, "$", "")
		case "currency":
			receipt.Currency = value
		case "merchant":
			receipt.Merchant = value
		case "category":
			receipt.Category = value
		case "description":
			receipt.Description = value
		case "receipt":
			receipt.IsReceipt = value == "true"
		case "items":
			receiptLines, err := parseReceiptLines(value)
			if err != nil {
				return Receipt{}, err
			}
			receipt.ReceiptLines = append(receipt.ReceiptLines, receiptLines...)
		case "first_event":
			receipt.IsFirstEvent = value == "true"
		case "title":
			receipt.Title = value
		case "name":
			receipt.Name = value
		case "parent_screen":
		case "place":
		case "price": // pretty sure this happens when a receipt line is updated
		case "quantity": // same as above
		case "rs_subscription":
		case "export_format":
		case "success":
		case "app_install_time":
		case "plan":
		case "method":
		case "sort_by":
		case "product":
		case "provider":
		case "tags":
		case "user_purpose":
		case "purchase_id":
		case "onboarding_version":
		case "referrer_click_time":
		case "type":
		case "offer":
		case "receipt_attached":
		case "utm_source":
		case "utm_medium":
		case "receipts_count":
			// ignore for now. maybe support in future
			continue
		default:
			return Receipt{}, fmt.Errorf("unknown key: %s", key)
		}
	}
	receipt.IsReconciled = checkIsReconciled(receipt)
	return receipt, nil
}

func parseCleanProps(logLine string) string {
	logLine = strings.ReplaceAll(logLine, "\r", "")
	logLine = strings.ReplaceAll(logLine, "\n", "")
	re := regexp.MustCompile(`props:\s*{`)
	matches := re.FindStringIndex(logLine)
	if matches == nil {
		return ""
	}
	propsStart := matches[1]
	props := logLine[propsStart:]
	props = strings.TrimRight(props, "}")
	return props
}

func parseUnixtime(unixtime string) string {
	// chop off the last 3 digits
	unixtime = unixtime[:len(unixtime)-3]
	unixtimeInt, err := strconv.ParseInt(unixtime, 10, 64)
	if err != nil {
		fmt.Println("Error parsing Unix time:", err)
		return ""
	}

	t := time.Unix(unixtimeInt, 0)
	formattedDate := t.Format("2006-01-02") // YYYY-MM-DD
	return formattedDate
}

func parseReceiptLines(items string) ([]ReceiptLine, error) {
	// 1_Item
	//     name: WELLNESS ROAD LINSEE 500GRAM 3 @ $3.30 EACH,
	//     quantity: 0.0,
	//     pricePerUnit: ,
	//     totalPrice: 990
	// 2_Item
	//     name: % TULIPS 1EACH,
	//     quantity: 0.0,
	//     pricePerUnit: ,
	//     totalPrice: 2000,

	receiptLines := []ReceiptLine{}

	receiptLineStrings := strings.Split(items, "_Item")
	numReceiptLineStrings := len(receiptLineStrings)
	for i, receiptLineString := range receiptLineStrings {
		if i == 0 && receiptLineString == "1" {
			continue
		}
		if receiptLineString == "" {
			continue
		}
		isLastItem := i == (numReceiptLineStrings - 1)
		shouldRemoveItemNumber := (i > 0) && !isLastItem

		if shouldRemoveItemNumber {
			itemNumber := i + 1
			// remove the last character from the string
			receiptLineString = strings.TrimRight(
				receiptLineString, strconv.Itoa(itemNumber))
		}

		receiptLine := ReceiptLine{}

		keyValuePairs := strings.Split(receiptLineString, ",")
		for _, keyValuePairString := range keyValuePairs {
			keyValuePair := strings.Split(strings.TrimSpace(keyValuePairString), ":")
			key := strings.TrimSpace(keyValuePair[0])
			value := strings.TrimSpace(keyValuePair[1])
			switch key {
			case "name":
				receiptLine.Name = value
			case "quantity":
				receiptLine.Quantity = value
			case "pricePerUnit":
				pricePerUnit, err := parsePrice(value)
				if err != nil {
					return receiptLines, err
				}
				receiptLine.PricePerUnit = pricePerUnit
			case "totalPrice":
				totalPrice, err := parsePrice(value)
				if err != nil {
					return receiptLines, err
				}
				receiptLine.TotalPrice = totalPrice
			default:
				return nil, fmt.Errorf("unknown key: %s", key)
			}
		}

		receiptLines = append(receiptLines, receiptLine)
	}

	return receiptLines, nil
}

func parsePrice(price string) (string, error) {
	price = strings.TrimSpace(price)
	price = strings.ReplaceAll(price, "$", "")
	if price == "" {
		return "0.00", nil
	}
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return "", err
	}
	priceFloat = priceFloat / 100
	return fmt.Sprintf("%.2f", priceFloat), nil
}

func checkIsReconciled(receipt Receipt) bool {
	runningTotal := 0.0
	for _, line := range receipt.ReceiptLines {
		if line.TotalPrice == "" {
			return false
		}
		parsedRunningTotal, _ := strconv.ParseFloat(line.TotalPrice, 64)
		runningTotal += parsedRunningTotal
	}
	parsedTotal, _ := strconv.ParseFloat(receipt.Total, 64)
	return runningTotal == parsedTotal
}

func toCSV(receipts []Receipt) string {
	var builder strings.Builder
	builder.WriteString("LogLine,Date,Title,Name,Total,Currency,Merchant,Category," +
		"Description,IsReconciled,ItemName,Quantity,PricePerUnit,TotalPrice\n")

	for _, receipt := range receipts {
		for _, line := range receipt.ReceiptLines {
			builder.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%t,%s,%s,%s,%s\n",
				receipt.LineNumber,
				receipt.Date,
				receipt.Title,
				receipt.Name,
				receipt.Total,
				receipt.Currency,
				receipt.Merchant,
				receipt.Category,
				receipt.Description,
				receipt.IsReconciled,
				line.Name,
				line.Quantity,
				line.PricePerUnit,
				line.TotalPrice))
		}
	}
	return builder.String()
}

func writeToFile(data string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// loop through all lines of data and write to file
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}
}
