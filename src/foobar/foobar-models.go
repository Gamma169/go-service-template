package main

import (
	"database/sql"
	"errors"
	"github.com/Gamma169/go-server-helpers/db"
	"github.com/Gamma169/go-server-helpers/server"
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
	TempID string `json:"__id__" jsonapi:"attr,__id__"`

	SubModels []*SubModel `json:"subModels" jsonapi:"relation,sub-models"`
}

func (f *FoobarModel) Validate() (err error) {

	if strings.TrimSpace(f.Name) == "" {
		err = errors.New("Cannot have empty name")
	} else if f.SomeProp == "bad prop" {
		err = errors.New("SomeProp cannot equal 'bad prop'")
	} else {
		err = db.CheckStructFieldsForInjection(*f)
	}

	for _, subModel := range f.SubModels {
		if f.Id != "" && f.Id != *subModel.FoobarModelId {
			err = errors.New("Cannot assign subModel to another model")
			return
		}
		if err = subModel.Validate(); err != nil {
			return
		}
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

	err = db.AssignArrayPropertyFromString(f, "SomeArrProp", arrStr, DB_ARRAY_DELIMITER)

	// We perform the same function many times- so instead of checking the err every time,
	// we can do this wrapping to stop execution if any of them throw an error
	// From: https://stackoverflow.com/questions/15397419/go-handling-multiple-errors-elegantly
	// assignPropWrapper := func(propStr string, valStr string) bool {
	//     err = db.AssignArrayPropertyFromString(f, propStr, valStr)
	//     return err == nil
	// }

	// need to assign to an _ because otherwise I get value is 'evaluated but not used' from compiler
	// _ = assignPropWrapper("p1", p1s) &&
	// assignPropWrapper("p2", p2s) &&
	// assignPropWrapper("p3", p3s)

	return
}

func (f *FoobarModel) ConvertToDatabaseInput(requesterId string) []interface{} {
	if f.Id == ZERO_UUID || strings.TrimSpace(f.Id) == "" {
		f.Id = uuid.New().String()
		for _, subModel := range f.SubModels {
			subModel.FoobarModelId = &f.Id
		}
	}
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

func initFoobarModelsPreparedStatements() {
	var err error

	if getFoobarModelsStmt, err = DB.Prepare(`
		SELECT f.id, 
			f.name, f.age, f.some_prop, f.some_nullable_prop, f.some_arr_prop,
			f.date_created, f.last_updated 
		FROM foobar_models f
		WHERE f.user_id = ($1);
	`); err != nil {
		panic(err)
	}

	if postFoobarModelStmt, err = DB.Prepare(`
		INSERT INTO foobar_models (
			name, age, some_prop, some_nullable_prop, some_arr_prop,
			date_created, last_updated,
			id, user_id
		)
		VALUES (($1), ($2), ($3), ($4), ($5), ($6), ($7), ($8), ($9))
	`); err != nil {
		panic(err)
	}

	if updateFoobarStmt, err = DB.Prepare(`
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
	`); err != nil {
		panic(err)
	}

	if deleteFoobarStmt, err = DB.Prepare(`
		DELETE FROM foobar_models WHERE id = ($1) AND user_id = ($2);
	`); err != nil {
		panic(err)
	}

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
	defer func() { server.SendErrorOnError(err, http.StatusInternalServerError, w, r, logError) }()

	requesterId := r.Header.Get(REQUESTER_ID_HEADER) // Should exist and be valid because of middleware

	var foobarModels []*FoobarModel
	if foobarModels, err = getModelsForRequester(requesterId); err != nil {
		return
	}

	if err = server.WriteModelToResponseFromHeaders(foobarModels, http.StatusOK, w, r); err == nil {
		debugLog("Found Models ≡(*′▽`)っ!")
	}
}

func CreateOrUpdateFoobarModelHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("Got Request To Create Or Update Model")
	var err error
	var errStatus = http.StatusInternalServerError

	defer func() { server.SendErrorOnError(err, errStatus, w, r, logError) }()

	// Need to initialize array or json response will be null
	model := FoobarModel{
		SubModels: []*SubModel{},
	}
	if err = server.PreProcessInputFromHeaders(&model, 32768, w, r); err != nil {
		errStatus = http.StatusBadRequest
		return
	}

	respStatus := http.StatusOK
	requesterId := r.Header.Get(REQUESTER_ID_HEADER) // Should exist and be valid because of middleware
	if r.Method == http.MethodPost {
		respStatus = http.StatusCreated
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

	if err = server.WriteModelToResponseFromHeaders(model, respStatus, w, r); err == nil {
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
	modelsMap := map[string]*FoobarModel{}

	for fbRows.Next() {
		// Need to initialze the array or json response will return it as null
		model := FoobarModel{
			SubModels: []*SubModel{},
		}

		if err = model.ScanFromRowsOrRow(fbRows); err != nil {
			return nil, err
		}

		foobarModels = append(foobarModels, &model)
		modelsMap[model.Id] = &model
	}

	subModels, err := getSubModelsForRequester(requesterId)
	if err != nil {
		return nil, err
	}

	for _, subModel := range subModels {
		if subModel.FoobarModelId != nil {
			model, ok := modelsMap[*subModel.FoobarModelId]
			if ok {
				model.SubModels = append(model.SubModels, subModel)
			}
		}
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

	if len(model.SubModels) > 0 {
		if err := postSubModelsToDatabase(model.SubModels, model.Id, requesterId, update); err != nil {
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
	// Note we do not need to delete the sub_models because of the `cascade` foreign key property in the migrations
	// If you delete its foregn kety, anny associated sub_models will also be deleted.
	return
}
