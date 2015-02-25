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
)

type Stock struct {
	XMLName xml.Name `xml:"query"`
	Data []Data `xml:"results>row"`
}

type Data struct {
	XMLName xml.Name `xml:"row"`
	Symbol string `xml:"symbol"`
	LastTradeDate string `xml:"lastTradeDate"`
	LastTradeTime string `xml:"lastTradeTime"`
	LastTradePrice string `xml:"lastTradePrice"`
	Change string `xml:"change"`
	ChangePct string `xml:"changePct"`
	Open string `xml:"open"`
	High string `xml:"high"`
	Low string `xml:"low"`
}


func main() {
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

	resp, err := http.Get("https://query.yahooapis.com/v1/public/yql?" + v.Encode())

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	decodeXml(resp.Body)

}

func substr(s string, pos, length int) string {
    runes:=[]rune(s)
    l := pos+length
    if l > len(runes) {
        l = len(runes)
    }
    return string(runes[pos:l])
}

func (d Data) String() string {
	color := 0
	change, err := strconv.ParseFloat(d.Change, 32)
	if err != nil {
		log.Fatal(err)
	}
	if change > 0 {
		color = 32
	} else {
		color = 31
	}
	changePct := strings.Split(d.ChangePct, " ")
	percent := changePct[len(changePct)-1]
	return fmt.Sprintf("\033[1m %s  \t\033[0m\033[%dm%s (%s)\t\033[0m%s\t%s %s\t%s\t%s\t%s", d.Symbol,color, d.Change, percent, d.LastTradePrice, d.LastTradeDate, d.LastTradeTime, d.Open, d.High, d.Low)

}

func decodeXml(body io.ReadCloser) {

	XMLdata := xml.NewDecoder(body)

	var s Stock

	err := XMLdata.Decode(&s)
	if err != nil {
		log.Fatal(err)
	}

	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	fmt.Fprintln(w, "\033[4mSymbol\tChange\033[4m\tLast\033[4m\t \tOpen\tHigh\tLow\033[0m")

	for _, stock := range s.Data {
		fmt.Fprintln(w, stock)
	}
	w.Flush()
}
