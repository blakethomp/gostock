package main

import (
	"fmt"
	"io"
	"io/ioutil"
  "net/http"
	"log"
	"encoding/csv"
	"strings"
	"strconv"
)

func main() {
	content, err := ioutil.ReadFile("stocks.txt")

	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Join(strings.Split(string(content), "\n"), ",")

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


	resp, err := http.Get("http://download.finance.yahoo.com/d/quote.csv?e=.csv&f=" + strings.Join(format, "") + "&s=" + lines)
	if err != nil {
		log.Fatal(err)
		fmt.Println("hi\n\n");
	}

	defer resp.Body.Close()

	parseCsv(resp.Body)

}

func parseCsv(body io.ReadCloser) {

	reader := csv.NewReader(body)
	record, err := reader.ReadAll()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Symbol	Date/Time		Last Trade		Change			Open		High		Low")
	for _, each := range record {
		color := 0
		change, err := strconv.ParseFloat(each[4], 32)
		if err != nil {
			log.Fatal(err)
		}
		if change > 0 {
			color = 32
		} else {
			color = 31
		}
		pct := strings.Split(each[5], " ")
		pt := pct[len(pct)-1]
		fmt.Printf("%s	%s %s	%s			\x1b[%dm%s (%s)\x1b[0m		%s		%s		%s\n", each[0], each[1], each[2], each[3], color, each[4], pt, each[5], each[6], each[7])
	}
}

func substr(s string, pos, length int) string{
    runes:=[]rune(s)
    l := pos+length
    if l > len(runes) {
        l = len(runes)
    }
    return string(runes[pos:l])
}
// uses encoding/json to decode a json response
// func decode(resp []byte) {
// 	type Message struct {
// 		LastPrice, ChangePercent float32
// 		Symbol string
// 	}
// 	var m Message
// 	err := json.Unmarshal(resp, &m)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var color int
// 	if m.ChangePercent > 0 {
// 		color = 32
// 	} else {
// 		color = 31
// 	}
// 	fmt.Printf("%s: %f \x1b[%d;1m(%f)\x1b[0m\n", m.Symbol, m.LastPrice, color, m.ChangePercent)
// }
