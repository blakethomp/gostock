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
)

type Stock struct {
	XMLName xml.Name `xml:"query"`
	Data []Data `xml:"results>row"`
}

type Data struct {
	Symbol string `xml:"symbol"`
	Open float32 `xml:"open"`
	High float32 `xml:"high"`
	Low float32 `xml:"low"`
	Date string `xml:"lastTradeDate"`
	Time string `xml:"lastTradeTime"`
	Last float32 `xml:"lastTradePrice"`
	Change float32 `xml:"change"`
	Pct string `xml:"changePct"`
}

var clear bool

var intervalFlag time.Duration

func init() {
	flag.DurationVar(&intervalFlag, "interval", 3*time.Second, "interval to refresh list")
	flag.DurationVar(&intervalFlag, "i", 3*time.Second, "interval to refresh list")
}

func main() {
	clear = false

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

	if clear {
		fmt.Fprintf(w, "\033[%dA", len(s.Data) + 2)
	} else {
		// Clear the terminal and reset the cursor
		fmt.Print("\033[2J\033[H")
		clear = true
	}

	fmt.Fprintln(w, "\033[K" + time.Now().Round(time.Second).String())

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

	// If the change is positive make it green, else red
	if d.Change > 0 {
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
				flt := v.Field(i).Float()
				value = strconv.FormatFloat(flt, 'f', 2, 32)

				if d.Change > 0 {
					value = "+" + value
				}

			case "Pct":
				ansi = color

			case "Symbol":
				ansi = "1"

			case "Open", "High", "Low", "Last":
				ansi = "0"
				value = strconv.FormatFloat(v.Field(i).Float(), 'f', 2, 32)

			default:
				ansi = "0"
		}


		fs = "\033[%sm%s\033[0m\t"
		s += fmt.Sprintf(fs, ansi, value)
	}

	return s
}
