package main

import (
    "errors"
    "fmt"
    "database/sql"
    _ "github.com/lib/pq"
    "reflect"
    "strings"
    "time"
)


func initDB() {
    debugLog("Establishing connection with database")
    var err error

    dbURL := getOptionalEnv("DATABASE_URL", "") 
    if dbURL != "" {
        DB, err = sql.Open("postgres", dbURL)    
    } else {
        DB, err = sql.Open("postgres", 
            fmt.Sprintf("user='%s' password='%s' dbname='%s' host='%s' port=%s sslmode=%s", 
                getRequiredEnv("DATABASE_USER"), 
                getOptionalEnv("DATABASE_PASSWORD", ""), 
                getRequiredEnv("DATABASE_NAME"), 
                getRequiredEnv("DATABASE_HOST"), 
                getOptionalEnv("DATABASE_PORT", "5432"), 
                getOptionalEnv("SSL_MODE", "disable")))
    }

    if (err != nil) {
        logError(errors.New("Error with sql Open statement"), nil)
        logError(err, nil)
        panic(err)
    }

    for tries := 0; tries == 0 || err != nil; tries++ {
        err = DB.Ping()
            
        if err != nil {
            if tries > 2 {
                logError(errors.New("Error: Could not connect to DB"), nil)
                panic(err)
            }
            debugLog("Error: Could not connect to DB -- trying again in 3 seconds")
            time.Sleep(time.Duration(3)*time.Second)
        }
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

// Function mostly from
// https://coderedirect.com/questions/432349/golang-dynamic-access-to-a-struct-property
func assignArrayPropertyFromString(st interface{}, field string, arrString string) error {
    // st must be a pointer to a struct
    refSt := reflect.ValueOf(st)
    if refSt.Kind() != reflect.Ptr || refSt.Elem().Kind() != reflect.Struct {
        return errors.New("st must be pointer to struct")
    }

    // Dereference pointer
    refSt = refSt.Elem()

    // Lookup field by name
    fieldSt := refSt.FieldByName(field)
    if !fieldSt.IsValid() {
        return fmt.Errorf("not a field name: %s", field)
    }

    // Field must be exported
    if !fieldSt.CanSet() {
        return fmt.Errorf("cannot set field %s", field)
    }

    // We expect an array field
    if fieldSt.Kind() != reflect.Slice && fieldSt.Kind() != reflect.Array {
        debugLog(fieldSt.Kind())
        return fmt.Errorf("%s is not a slice or array field", field)
    }

    arr := []string{}
    if arrString != "" {
        arr = strings.Split(arrString, DB_ARRAY_DELIMITER)
    }

    fieldSt.Set(reflect.ValueOf(arr))
    return nil
}
