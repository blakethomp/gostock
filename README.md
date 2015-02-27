# gostock
Command line stock ticker in golang

Retrieves stock data from Yahoo and prints it in a tabulated display. Yahoo limits public requests to 2,000 requests/hour per IP.

Stock symbols are read from stocks.txt, one symbol on each line

This will refresh at a default interval of 3 seconds, but this can be set using the interval flag:

```
-i 5s
-interval 5s
```

A stocks.txt file will need to be placed in your user directory

