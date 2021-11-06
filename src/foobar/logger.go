package main

import (
    "fmt"
    "net/http"
)

func logError(err error, r *http.Request) {
    if debug {
        fmt.Println(err)
    } else {
        fmt.Println(err)
    }
}

func debugLog(msg interface{}) {
    if debug {
        fmt.Println(msg)
    }
}
