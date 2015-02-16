package main

import (
	"fmt"
	"io/ioutil"
  	"net/http"
	"log"
	"os"
	"encoding/json"
)

func main() {
	fmt.Println(os.Args[1])
	res, err := http.Get("http://dev.markitondemand.com/Api/v2/Quote/json?symbol="+os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	resp, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", resp)
	decode(resp)
}

func decode(resp []byte) {
	type Message struct {
		LastPrice, ChangePercent float32
		Symbol string
	}
	var m Message
	err := json.Unmarshal(resp, &m)
	if err != nil {
		log.Fatal(err)
	}
	var color int
	if m.ChangePercent > 0 { 
		color = 32 
	} else { 
		color = 31 
	}
	fmt.Printf("%s: %f \x1b[%d;1m(%f)\x1b[0m\n", m.Symbol, m.LastPrice, color, m.ChangePercent)
}
		
