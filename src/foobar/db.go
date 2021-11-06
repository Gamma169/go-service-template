package main

import (
    "errors"
    "fmt"
    "database/sql"
    _ "github.com/lib/pq"
    "reflect"
    "strings"
)


func initDB() {
    debugLog("Establishing connection with database")
    var err error

    DB, err = sql.Open("postgres", 
        fmt.Sprintf("user='%s' password='%s' dbname='%s' host='%s' port=%s sslmode=%s", 
            getRequiredEnv("DATABASE_USER"), 
            getOptionalEnv("DATABASE_PASSWORD", ""), 
            getRequiredEnv("DATABASE_NAME"), 
            getRequiredEnv("DATABASE_HOST"), 
            getOptionalEnv("DATABASE_PORT", "5432"), 
            getOptionalEnv("SSL_MODE", "disable")))

    if (err != nil) {
        logError(errors.New("Error with sql Open statement"), nil)
        logError(err, nil)
        panic(err)
    }

    if err = DB.Ping(); err != nil {
        logError(errors.New("Error: Could not connect to DB"), nil)
        logError(err, nil)
        panic(err)
    }
    debugLog("Connection sucessfully established")    
}


// Generally this function isn't necessary because we use prepared statements, but more safety is good
func checkStructFieldsForInjection(st interface{}) error {
    t := reflect.TypeOf(st)
    for i := 0; i < t.NumField(); i++ {
        if t.Field(i).Type.Kind() == reflect.String {
            r := reflect.ValueOf(st)
            if strings.Contains(reflect.Indirect(r).Field(i).String(), ";") {
                return errors.New("Found semicolon in string struct field")
            }
        }
    }
    return nil
}
