package main

import (
	"database/sql"
	"errors"
	// "fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"time"
)

type FoobarModel struct {
	Id               string   `json:"id" jsonapi:"primary,foobar-model"`
	Name             string   `json:"name" jsonapi:"attr,name"`
	Age              int      `json:"age" jsonapi:"attr,age"`
	SomeProp         string   `json:"someProp" jsonapi:"attr,some-prop"`
	SomeNullableProp *string  `json:"someNullableProp" jsonapi:"attr,some-nullable-prop"`
	SomeArrProp      []string `json:"someArrProp" jsonapi:"attr,some-arr-prop"`

	DateCreated time.Time `json:"dateCreated" jsonapi:"attr,date-created"`
	LastUpdated time.Time `json:"lastUpdated" jsonapi:"attr,last-updated"`

	// Only needed for jsonapi save-relationships-mixin
	TempID string `jsonapi:"attr,__id__"`

	// TODO
	// SubModels         []*SubModel  `jsonapi:"relation,sub-models"`
}

func (f *FoobarModel) Validate() (err error) {

	if strings.TrimSpace(f.Name) == "" {
		err = errors.New("Cannot have empty name")
	} else if f.SomeProp == "bad prop" {
		err = errors.New("SomeProp cannot equal 'bad prop'")
	} else {
		err = checkStructFieldsForInjection(*f)
	}

	return
}

func (f *FoobarModel) ScanFromRowsOrRow(rowsOrRow interface {
	Scan(dest ...interface{}) error
}) (err error) {
	// Define any vars that don't fit in the model
	var (
		nullableStr sql.NullString
		arrStr      string
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

	// need to assign to an _ because otherwise I get value is 'evaluated but not used' from compiler
	// _ = assignPropWrapper("p1", p1s) &&
	// assignPropWrapper("p2", p2s) &&
	// assignPropWrapper("p3", p3s)

	return
}

func (f *FoobarModel) ConvertToDatabaseInput(requesterId string) []interface{} {
	f.LastUpdated = time.Now()
	return []interface{}{
		f.Name,
		f.Age,
		f.SomeProp,
		f.SomeNullableProp,
		strings.Join(f.SomeArrProp, DB_ARRAY_DELIMITER),

		f.DateCreated,
		f.LastUpdated,

		f.Id,
		requesterId,
	}
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
            name, age, some_prop, some_nullable_prop, some_arr_prop,
            date_created, last_updated,
            id, user_id
        )
        VALUES (($1), ($2), ($3), ($4), ($5), ($6), ($7), ($8), ($9))
    `)

	updateFoobarStmt, err = DB.Prepare(`
        UPDATE foobar_models
        SET name = ($1),
            age = ($2),
            some_prop = ($3),
            some_nullable_prop = ($4),
            some_arr_prop = ($5),
            date_created = ($6),
            last_updated = ($7)
        WHERE
            id = ($8) AND
            user_id = ($9);
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
	defer func() { sendErrorOnError(err, http.StatusInternalServerError, w, r) }()

	requesterId := r.Header.Get(REQUESTER_ID_HEADER) // Should exist and be valid because of middleware

	var foobarModels []*FoobarModel
	if foobarModels, err = getModelsForRequester(requesterId); err != nil {
		return
	}

	if err = WriteModelToResponse(foobarModels, w, r); err == nil {
		debugLog("Found Models ≡(*′▽`)っ!")
	}
}

func CreateOrUpdateFoobarModelHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("Got Request To Create Or Update Model")
	var err error
	var errStatus = http.StatusInternalServerError

	defer func() { sendErrorOnError(err, errStatus, w, r) }()

	var model FoobarModel	
	if err, errStatus = PreProcessInput(&model, w, r, 32768); err != nil {
		return
	}

	requesterId := r.Header.Get(REQUESTER_ID_HEADER) // Should exist and be valid because of middleware
	if r.Method == http.MethodPost {
		if model.Id == ZERO_UUID || strings.TrimSpace(model.Id) == "" {
			model.Id = uuid.New().String()
		}
		model.DateCreated = time.Now()
		if err = postModelToDatabase(&model, requesterId, false); err != nil {
			return
		}
	} else if r.Method == http.MethodPatch {
		if err = postModelToDatabase(&model, requesterId, true); err != nil {
			return
		}
	} else {
		err = errors.New("Somehow calling create/update handler with not POST or PATCH request")
		return
	}

	if err = WriteModelToResponse(model, w, r); err == nil {
		debugLog("Created or Updated Model!")
	}
}

func DeleteFoobarModelHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("Received request to delete model")

	vars := mux.Vars(r)
	id := vars["modelId"]
	requesterId := r.Header.Get(REQUESTER_ID_HEADER) // Should exist and be valid because of middleware
	if err := deleteModelForRequester(id, requesterId); err != nil {
		debugLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	debugLog("Deleted Model! ٩(˘◡˘)۶")
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

func postModelToDatabase(model *FoobarModel, requesterId string, update bool) error {
	dbInput := model.ConvertToDatabaseInput(requesterId)

	if update {
		if _, err := updateFoobarStmt.Exec(dbInput...); err != nil {
			return err
		}
	} else {
		if _, err := postFoobarModelStmt.Exec(dbInput...); err != nil {
			return err
		}
	}

	return nil
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
