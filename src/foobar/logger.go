package main

import (
    "fmt"
    "net/http"
)

const BoldPrint = "\033[1m"
const HeaderPrint = "\033[95m"
const RedPrint = "\033[91m"
const BluePrint = "\033[94m"
const EndPrint = "\033[0m"

func logError(err error, r *http.Request) {
    traceId := "startup"
    if r != nil {
        traceId = r.Header.Get(TRACE_ID_HEADER)
    }
    if debug {
        fmt.Println(RedPrint + "Error: " + err.Error() + " -- " + traceId + EndPrint)
    } else {
        fmt.Println(RedPrint + "Error: " + err.Error() + " -- " + traceId + EndPrint)
    }
}

func debugLog(msg ...interface{}) {
    if debug {
        fmt.Println(msg, " ")
    }
}
