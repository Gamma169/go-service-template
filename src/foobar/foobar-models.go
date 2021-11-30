package main

import (
    // "database/sql"
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
    Id                   string     `json:"id", jsonapi:"primary,foobar-model"`
    Name                 string     `json:"name" jsonapi:"attr,name"`
    Age                  int        `json:"age" jsonapi:"attr,age"`
    SomeProp             string     `json:"someProp" jsonapi:"attr,some-prop"`
    SomeNullableProp     *string    `json:"someNullableProp" jsonapi:"attr,some-nullable-prop"`
    
    DateCreated          time.Time  `json:"dateCreated" jsonapi:"attr,date-created"`
    LastUpdated          time.Time  `json:"lastUpdated" jsonapi:"attr,last-updated"`

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

    getFoobarModelStmt, err = DB.Prepare(`
        SELECT f.id, 
            f.name, f.age, f.some_prop, f.some_nullable_prop,
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
            name, age, some_prop, some_nullable_prop,
            date_created, last_updated
        )
        VALUES (($1), ($2), ($3), ($4), ($5), ($6), ($7), ($8))
    `)
    

    updateFoobarStmt, err = DB.Prepare(`
        UPDATE foobar_models
        SET user_id = ($1),
            name = ($2),
            age = ($3),
            some_prop = ($4),
            some_nullable_prop = ($5)
        WHERE
            id = ($6);
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
    debugLog("Received request to get FoobarModels for user")

    // Should exist and be valid because of middleware
    userId := r.Header.Get("user-id")

    foobarModels, err, errStatus := getFoobarModelsForUser(userId)
    if err != nil {
        debugLog(err)
        http.Error(w, err.Error(), errStatus)
        return
    }

    w.Header().Set("Content-Type", jsonapi.MediaType)
    if err := jsonapi.MarshalPayload(w, foobarModels); err != nil {
        debugLog(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    debugLog("Found FoobarModels!")
}


/*********************************************
 * Database Functions
 * *******************************************/

// TODO
func getFoobarModelsForUser(userId string) ([]*FoobarModel, error, int) {
    _, err := getFoobarModelStmt.Query(userId)
    if err != nil {
        return nil, err, http.StatusInternalServerError
    }
    return nil, nil, 0
}


