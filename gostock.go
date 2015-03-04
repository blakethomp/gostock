package main

import (
	"fmt"
	"os"
	"os/user"
	"text/tabwriter"
	"io"
	"io/ioutil"
	"net/http"
	"log"
	"encoding/xml"
	"strings"
	"strconv"
	"net/url"
	"reflect"
	"time"
	"flag"
	"regexp"
)

type Stock struct {
	XMLName xml.Name `xml:"query"`
	Data []Data `xml:"results>row"`
}

var stockContainer Stock

type Data struct {
	Symbol string `xml:"symbol"`
	Open string `xml:"open"`
	High string `xml:"high"`
	Low string `xml:"low"`
	Date string `xml:"lastTradeDate"`
	Time string `xml:"lastTradeTime"`
	Last string `xml:"lastTradePrice"`
	Change string `xml:"change"`
	Pct string `xml:"changePct"`
}

var intervalFlag time.Duration

func init() {
	flag.DurationVar(&intervalFlag, "interval", 3*time.Second, "interval to refresh list")
	flag.DurationVar(&intervalFlag, "i", 3*time.Second, "interval to refresh list")
}

func main() {

	flag.Parse()

	duration := intervalFlag

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile(usr.HomeDir + "/stocks.txt")

	if err != nil {
		log.Fatal(err)
	}

	loadData(data)

	for _ = range time.Tick(duration) {
		loadData(data)
	}
}

func loadData(data []byte) {

	symbols := strings.Join(strings.Split(string(data), "\n"), ",")

	format := make([]string, 9)
	format[0] = "s" 	// Symbol
	format[1] = "d1" 	// Last Trade Date
	format[2] = "t1"	// Last Trade Time
	format[3] = "l1"	// Last Trade (Price Only)
	format[4] = "c6"	// Change (Realtime)
	format[5] = "p2"	// Change Percent (Realtime)
	format[6] = "o"		// Open
	format[7] = "h"		// Day's High
	format[8] = "g"		// Day's Low

	v := url.Values{}

	v.Set("q", "select * from csv where url='http://download.finance.yahoo.com/d/quotes.csv?f=" + strings.Join(format, "") + "&s=" + symbols + "&e=.csv' and columns='symbol,lastTradeDate,lastTradeTime,lastTradePrice,change,changePct,open,high,low'")
	v.Add("format", "xml")
	v.Add("env", "store://datatables.org/alltableswithkeys")


	resp, err := http.Get("https://query.yahooapis.com/v1/public/yql?" + v.Encode())

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close() // Defer the closing of the request

	stocks := decodeXml(resp.Body)

	formatOutput(stocks)
}

func decodeXml(body io.ReadCloser) Stock {

	XMLdata := xml.NewDecoder(body)

	// reset value of stockContainer
	stockContainer = Stock{}

	err := XMLdata.Decode(&stockContainer)
	if err != nil {
		log.Fatal(err)
	}

	return stockContainer
}

func formatOutput (s Stock) {
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	fmt.Print("\033[2J\033[H")

	fmt.Fprintln(w, time.Now().Round(time.Second).String())

	var d Data
	v := reflect.ValueOf(d) // reflect lets us iterate on the struct

	var value, separator, header string

	for i := 0; i < v.NumField(); i++ {
		value = v.Type().Field(i).Name
		if (i < (v.NumField() - 1)) {
			separator = "\t"
		} else {
			separator = ""
		}

		// Print the header labels underlined
		header += fmt.Sprintf("\033[4m%s\033[0m%s", value, separator)
	}

	fmt.Fprintln(w, header)

	// run the stock through String()
	for _, stock := range s.Data {
		fmt.Fprintln(w, stock)
	}

	w.Flush()
}

func (d Data) String() string {

	v := reflect.ValueOf(d)

	var fs, s, ansi, value string

	// We're starting at 1 to skip the XML field name
	for i := 0; i < v.NumField(); i++ {
		value = v.Field(i).String()
		switch v.Type().Field(i).Name {

			case "Change":
				val, err := strconv.ParseFloat(value, 64)
				if err == nil {
					value = strconv.FormatFloat(val, 'f', 2, 64)
				}

				if val > 0 {
					ansi = "32"
					value = "+" + value
				} else {
					ansi = "31"
				}

			case "Pct":
				val, err := strconv.ParseFloat(strings.Replace(value, "%", "", 1), 64)
				if err == nil {
					value = strconv.FormatFloat(val, 'f', 2, 64)
				}

				if val > 0 {
					ansi = "32"
					value = "+" + value
				} else {
					ansi = "31"
				}

				value += "%"

			case "Symbol":
				ansi = "1"
				maxLen := maxSymbolLength(stockContainer)
				if len(value) < maxLen {
					diff := maxLen - len(value)
					value += strings.Repeat(" ", diff)
				}

			case "Open", "High", "Low", "Last":
				ansi = "0"
				val, err := strconv.ParseFloat(value, 64)
				if err == nil {
					value = strconv.FormatFloat(val, 'f', 2, 64)
				}

			case "Date":
				reg, err := regexp.Compile("\\d{2}(\\d{2})$")
				if err == nil {
					value = reg.ReplaceAllString(value, "$1")
				}

			default:
				ansi = "0"
		}

		fs = "\033[%sm%s\033[0m\t"
		s += fmt.Sprintf(fs, ansi, value)
	}

	return s
}

func maxSymbolLength(s Stock) int {
	size := 0
	for _, stock := range s.Data {
		if len(stock.Symbol) > size {
			size = len(stock.Symbol)
		}
	}
	return size
}
