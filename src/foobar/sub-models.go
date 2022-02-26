package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Gamma169/go-server-helpers/db"
	"github.com/google/uuid"
	"strings"
)

type SubModel struct {
	Id            string  `json:"id" jsonapi:"primary,sub-models"`
	FoobarModelId *string `json:"foobarModelId" jsonapi:"attr,foobar-model-id"`
	Value         string  `json:"value" jsonapi:"attr,value"`
	ValueInt      int     `json:"valueInt" jsonapi:"attr,value-int"`
	TempID        string  `json:"__id__" jsonapi:"attr,__id__"`
}

func (s *SubModel) Validate() (err error) {
	if s.Value == "bad prop" {
		err = errors.New("Value cannot equal 'bad prop'")
	} else {
		err = db.CheckStructFieldsForInjection(*s)
	}

	return
}

func (s *SubModel) ScanFromRowsOrRow(rowsOrRow interface {
	Scan(dest ...interface{}) error
}) (err error) {
	// Define any vars that don't fit in the model
	var (
		foobarModelId sql.NullString
	)

	// Define properties to be passed into scan-- must match prepared statement query order!
	properties := []interface{}{
		&s.Id,
		&foobarModelId, &s.Value, &s.ValueInt,
	}

	if err = rowsOrRow.Scan(properties...); err != nil {
		return
	}

	// Post process fields that don't get filled on the model itself
	if foobarModelId.Valid {
		s.FoobarModelId = &foobarModelId.String
	}

	return
}

func (s *SubModel) ConvertToDatabaseInput(requesterId string) []interface{} {
	if s.Id == "" {
		s.Id = uuid.New().String()
	}
	return []interface{}{
		s.Id,
		requesterId,
		s.FoobarModelId,
		s.Value,
		s.ValueInt,
	}
}

func initSubModelPreparedStatements() {
	var err error

	if getSubModelsForUserStmt, err = DB.Prepare(`
        SELECT s.id,
            s.foobar_model_id, s.value, s.value_int
            FROM sub_models s
        WHERE s.user_id = ($1);
    `); err != nil {
		panic(err)
	}

	if deleteSubModelsForFoobarModelStmt, err = DB.Prepare(`
        DELETE FROM sub_models WHERE foobar_model_id = ($1);
    `); err != nil {
		panic(err)
	}

}

/*********************************************
 * Database Functions
 * *******************************************/

func getSubModelsForRequester(requesterId string) ([]*SubModel, error) {
	sRows, err := getSubModelsForUserStmt.Query(requesterId)
	if err != nil {
		return nil, err
	}
	defer sRows.Close()

	subModels := []*SubModel{}

	for sRows.Next() {
		subModel := SubModel{}

		if err = subModel.ScanFromRowsOrRow(sRows); err != nil {
			return nil, err
		}

		subModels = append(subModels, &subModel)
	}

	return subModels, nil
}

func postSubModelsToDatabase(subModels []*SubModel, foobarModelId string, requesterId string, deleteFirst bool) error {
	if deleteFirst {
		if err := deleteSubModelsForFoobarModel(foobarModelId); err != nil {
			return err
		}
	}

	var insertQuery strings.Builder
	insertQuery.WriteString(`
        INSERT INTO sub_models (id, user_id, foobar_model_id, value, value_int)
        VALUES
    `)

	var vals []interface{}

	for i, subModel := range subModels {

		insertQuery.WriteString(fmt.Sprintf("(($%d), ($%d), ($%d), ($%d), ($%d))",
			(i*5)+1, (i*5)+2, (i*5)+3, (i*5)+4, (i*5)+5))
		if i != len(subModels)-1 {
			insertQuery.WriteString(",")
		}

		dbInput := subModel.ConvertToDatabaseInput(requesterId)
		vals = append(vals, dbInput...)
	}
	insertQuery.WriteString(";")

	insertStmt, err := DB.Prepare(insertQuery.String())
	if err != nil {
		return err
	}
	if _, err = insertStmt.Exec(vals...); err != nil {
		return err
	}

	return nil
}

func deleteSubModelsForFoobarModel(foobarModelId string) error {
	if _, err := deleteSubModelsForFoobarModelStmt.Exec(foobarModelId); err != nil {
		return err
	}
	return nil
}
