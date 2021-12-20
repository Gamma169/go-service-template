package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    // "fmt"
    // "github.com/google/uuid"
    "github.com/gorilla/mux"
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

func (f *FoobarModel)Validate() (err error) {

    if strings.TrimSpace(f.Name) == "" {
        err = errors.New("Cannot have empty name")
    } else if f.SomeProp == "bad prop" {
        err = errors.New("SomeProp cannot equal 'bad prop'")
    } 

    return
}

func (f *FoobarModel)ScanFromRowsOrRow(rowsOrRow interface{Scan(dest ...interface{}) error}) (err error) {
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

    if err = rowsOrRow.Scan(properties...); err != nil {
        return
    }
    
    // Postprocess + assign the vars above
    if nullableStr.Valid {
        f.SomeNullableProp = &nullableStr.String
    }

    err = assignArrayPropertyFromString(f, "SomeArrProp", arrStr)
    
    // We perform the same function many times- so instead of checking the err every time, 
    // we can do this wrapping to stop execution if any of them throw an error
    // From: https://stackoverflow.com/questions/15397419/go-handling-multiple-errors-elegantly
    // assignPropWrapper := func(propStr string, valStr string) bool {
    //     err = assignArrayPropertyFromString(f, propStr, valStr)
    //     return err == nil
    // }

    // // need to wrap in an anonymous no-op func because otherwise I get value is 'evaluated but not used'
    // func(b bool) {} (
    //     assignPropWrapper("p1", p1s) &&
    //     assignPropWrapper("p2", p2s) &&
    //     assignPropWrapper("p3", p3s),
    // )

    return
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
        DELETE FROM foobar_models WHERE id = ($1) AND user_id = ($2);
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

    // We have to invoke an anonymous function function here because if we just call sendErrorOnError directly
    // then we pass in 'err' as it is first initialized-- nil.  And we lose error handling
    // By wrapping in the anonymous function, err gets passed in with its value at the end of the call as we expect.
    defer func() {sendErrorOnError(err, http.StatusInternalServerError, w, r)}()

    requesterId := r.Header.Get(REQUESTER_ID_HEADER)  // Should exist and be valid because of middleware

    var foobarModels []*FoobarModel
    if foobarModels, err = getModelsForRequester(requesterId); err != nil {
        return
    }

    header := r.Header.Get(ContentTypeHeader)
    acceptHeader := r.Header.Get(AcceptContentTypeHeader)
    if header == jsonapi.MediaType || acceptHeader == jsonapi.MediaType {
        w.Header().Set(ContentTypeHeader, jsonapi.MediaType)
        err = jsonapi.MarshalPayload(w, foobarModels)  // don't need to wrap in if/return stmt b/c this is last stmt in func
    } else {
        w.Header().Set(ContentTypeHeader, JSONContentType)
        err = json.NewEncoder(w).Encode(foobarModels)  // don't need to wrap in if/return stmt b/c this is last stmt in func
    }

    if err == nil {
        debugLog("Found Models ≡(*′▽`)っ!")
    }
}



func DeleteFoobarModelHandler(w http.ResponseWriter, r *http.Request) {
    debugLog("Received request to delete model")

    vars := mux.Vars(r)
    id := vars["modelId"]
    requesterId := r.Header.Get(REQUESTER_ID_HEADER)  // Should exist and be valid because of middleware
    if err := deleteModelForRequester(id, requesterId); err != nil {
        debugLog(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)   
    debugLog("Deleted analytics file! ٩(˘◡˘)۶")
}


/*********************************************
 * Database Functions
 * *******************************************/

func getModelsForRequester(requesterId string) ([]*FoobarModel, error) {
    fbRows, err := getFoobarModelsStmt.Query(requesterId)
    if err != nil {
        return nil, err
    }
    defer fbRows.Close()

    foobarModels := []*FoobarModel{}

    for fbRows.Next() {
        model := FoobarModel{}

        if err = model.ScanFromRowsOrRow(fbRows); err != nil {
            return nil, err
        }
        
        foobarModels = append(foobarModels, &model)
    }


    return foobarModels, nil
}


func deleteModelForRequester(fileId string, requesterId string) (err error) {
    // NOTE: the other return value is an *sql.Result struct
    // We can possibly check if we actually deleted a file using result.RowsAffected() == 1,
    // and throw an error if == 0 (meaning we tried to delete a file that didn't exist or didn't belong to requester)
    // But I think that is unnecessary for now.  Just wanted to log possibility for posterity
    // https://pkg.go.dev/database/sql#Result
    _, err = deleteFoobarStmt.Exec(fileId, requesterId)
    return
}