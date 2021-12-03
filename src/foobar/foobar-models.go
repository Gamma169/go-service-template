package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    // "fmt"
    // "github.com/google/uuid"
    // "github.com/gorilla/mux"
    "github.com/google/jsonapi"
    "net/http"
    "strings"
    "time"
)

type FoobarModel struct {
    Id                   string     `json:"id" jsonapi:"primary,foobar-model"`
    Name                 string     `json:"name" jsonapi:"attr,name"`
    Age                  int        `json:"age" jsonapi:"attr,age"`
    SomeProp             string     `json:"someProp" jsonapi:"attr,some-prop"`
    SomeNullableProp     *string    `json:"someNullableProp" jsonapi:"attr,some-nullable-prop"`
    SomeArrProp          []string   `json:"someArrProp" jsonapi:"attr,some-arr-prop"`
    
    DateCreated          time.Time  `json:"dateCreated" jsonapi:"attr,date-created"`
    LastUpdated          time.Time  `json:"lastUpdated" jsonapi:"attr,last-updated"`

    // Only needed for jsonapi save-relationships-mixin
    TempID               string     `jsonapi:"attr,__id__"`

    // TODO
    // SubModels         []*SubModel  `jsonapi:"relation,sub-models"`
}

func (f *FoobarModel)Validate() error {
    var err error

    if strings.TrimSpace(f.Name) == "" {
        err = errors.New("Cannot have empty name")
    } else if f.SomeProp == "bad prop" {
        err = errors.New("SomeProp cannot equal 'bad prop'")
    } 

    return err
}

func (f *FoobarModel)ScanFromRowsOrRow(rows *sql.Rows, row *sql.Row) error {
    // Define any vars that don't fit in the model
    var (
        nullableStr sql.NullString
        arrStr string
    )

    // Define properties to be passed into scan-- must match prepared statement query order!
    properties := []interface{}{
        &f.Id, 
        &f.Name, &f.Age, &f.SomeProp,
        &nullableStr, &arrStr,
        &f.DateCreated, &f.LastUpdated,
    }

    if rows != nil {
        if err := rows.Scan(properties...); err != nil {
            return err
        }
    } else if row != nil {
        if err := row.Scan(properties...); err != nil {
            return err
        }
    } else {
        return errors.New("did not provide rows or row to scan from")
    }
    
    // Postprocess + assign the vars above
    if nullableStr.Valid {
        f.SomeNullableProp = &nullableStr.String
    }

    if err := assignArrayPropertyFromString(f, "SomeArrProp", arrStr); err != nil {
        return err
    }

    return nil
}

// TODO
// type SubModel struct {
//     Id             string     `jsonapi:"primary,stress-test-cohorts"`
//     FoobarModelId  string     `jsonapi:"attr,stress-test-id"`
//     Type           string
//     Value          string
//     TempID         string     `jsonapi:"attr,__id__"`
// }

func initFoobarModelsPreparedStatements() {
    var err error

    panicOnError := func() {
        if err != nil {
            panic(err)
        }
    }
    defer panicOnError()

    getFoobarModelsStmt, err = DB.Prepare(`
        SELECT f.id, 
            f.name, f.age, f.some_prop, f.some_nullable_prop, f.some_arr_prop,
            f.date_created, f.last_updated 
        FROM foobar_models f
        WHERE f.user_id = ($1);
    `)


// TODO Can make one for all as well to avoid multiple networks calls
//     getSubmodelForFoobarModelStmt, err = DB.Prepare(`
//         SELECT 
//         FROM submodels s
//         WHERE s.foobar_model_id = ($1);
//     `)

    postFoobarModelStmt, err = DB.Prepare(`
        INSERT INTO foobar_models (
            id, user_id, 
            name, age, some_prop, some_nullable_prop, some_arr_prop,
            date_created, last_updated
        )
        VALUES (($1), ($2), ($3), ($4), ($5), ($6), ($7), ($8), ($9))
    `)
    

    updateFoobarStmt, err = DB.Prepare(`
        UPDATE foobar_models
        SET user_id = ($1),
            name = ($2),
            age = ($3),
            some_prop = ($4),
            some_nullable_prop = ($5)
            some_arr_prop = ($6)
        WHERE
            id = ($7);
    `)

    deleteFoobarStmt, err = DB.Prepare(`
        DELETE FROM foobar_models WHERE id = ($1);
    `)

// TODO
//     deleteSubmodelStmt, err = DB.Prepare(`
//         DELETE FROM submodels WHERE foobar_model_id = ($1);
//     `)

}

/*********************************************
 * Request Handlers
 * *******************************************/


func GetFoobarModelHandler(w http.ResponseWriter, r *http.Request) {
    debugLog("Received request to get models for requester")
    var err error
    var errStatus = http.StatusInternalServerError

    // We have to invoke an anonymous function function here because if we just call sendErrorOnError directly
    // then we pass in 'err' as it is first initialized-- nil.  And we lose error handling
    // By wrapping in the anonymous function, err gets passed in with its value at the end of the call as we expect.
    defer func() {sendErrorOnError(err, errStatus, w, r)}()

    // Should exist and be valid because of middleware
    requesterId := r.Header.Get(REQUESTER_ID_HEADER)

    var foobarModels []*FoobarModel
    if foobarModels, err, errStatus = getModelsForRequester(requesterId); err != nil {
        return
    }

    header := r.Header.Get(ContentTypeHeader)
    acceptHeader := r.Header.Get(AcceptContentTypeHeader)
    if header == jsonapi.MediaType || acceptHeader == jsonapi.MediaType {
        w.Header().Set(ContentTypeHeader, jsonapi.MediaType)
        if err = jsonapi.MarshalPayload(w, foobarModels); err != nil {
            return
        }
    } else {
        w.Header().Set(ContentTypeHeader, JSONContentType)
        if err = json.NewEncoder(w).Encode(foobarModels); err != nil {
            return
        }
    }

    debugLog("Found Models ≡(*′▽`)っ!")
}


/*********************************************
 * Database Functions
 * *******************************************/

func getModelsForRequester(requesterId string) ([]*FoobarModel, error, int) {
    fbRows, err := getFoobarModelsStmt.Query(requesterId)
    if err != nil {
        return nil, err, http.StatusInternalServerError
    }
    defer fbRows.Close()

    foobarModels := []*FoobarModel{}

    for fbRows.Next() {
        model := FoobarModel{}

        if err = model.ScanFromRowsOrRow(fbRows, nil); err != nil {
            return nil, err, http.StatusInternalServerError
        }
        
        foobarModels = append(foobarModels, &model)
    }


    return foobarModels, nil, 0
}


