package main

import (
	"fmt"
	"os"
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
)

type Stock struct {
	XMLName xml.Name `xml:"query"`
	Data []Data `xml:"results>row"`
}

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


func main() {

	loadData(2)

}

func loadData(n time.Duration) {
	content, err := ioutil.ReadFile("stocks.txt")

	if err != nil {
		log.Fatal(err)
	}

	symbols := strings.Join(strings.Split(string(content), "\n"), ",")

	format := make([]string, 9)
	format[0] = "s" 	// Symbol
	format[1] = "d1" 	// Last Trade Date
	format[2] = "t1"	// Last Trade Time
	format[3] = "l1"	// Last Trade (Price Only)
	format[4] = "c6"	// Change (Realtime)
	format[5] = "k2"	// Change Percent (Realtime)
	format[6] = "o"		// Open
	format[7] = "h"		// Day's High
	format[8] = "g"		// Day's Low

	v := url.Values{}

	v.Set("q", "select * from csv where url='http://download.finance.yahoo.com/d/quotes.csv?f=" + strings.Join(format, "") + "&s=" + symbols + "&e=.csv' and columns='symbol,lastTradeDate,lastTradeTime,lastTradePrice,change,changePct,open,high,low'")
	v.Add("format", "xml")
	v.Add("env", "store://datatables.org/alltableswithkeys")

	for _ = range time.Tick(n * time.Second) {
		resp, err := http.Get("https://query.yahooapis.com/v1/public/yql?" + v.Encode())

		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close() // Defer the closing of the request

		stocks := decodeXml(resp.Body)

		formatOutput(stocks)
	}
}

func decodeXml(body io.ReadCloser) Stock {

	XMLdata := xml.NewDecoder(body)

	var s Stock

	err := XMLdata.Decode(&s)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func formatOutput (s Stock) {
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	// Clear the terminal and reset the cursor, print the time
	fmt.Fprintln(w, "\033[2J\033[H"+time.Now().String())

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
	color := "0"

	// Parse the change value into a float to check it's direction
	change, err := strconv.ParseFloat(d.Change, 32)
	if err != nil {
		log.Fatal(err)
	}

	// If the change is positive make it green, else red
	if change > 0 {
		color = "32"
	} else {
		color = "31"
	}

	v := reflect.ValueOf(d)

	var fs, s, ansi, value string

	// We're starting at 1 to skip the XML field name
	for i := 0; i < v.NumField(); i++ {
		value = v.Field(i).String()
		switch v.Type().Field(i).Name {

			case "Change":
				ansi = color

			case "Pct":
				pct := strings.Split(value, " ")
				value = pct[len(pct)-1]
				ansi = color

			case "Symbol":
				ansi = "1"

			case "Low":
				ansi = "0"

			default:
				ansi = "0"
		}

		fs = "\033[%sm%s\033[0m\t"
		s += fmt.Sprintf(fs, ansi, value)
	}

	return s
}
