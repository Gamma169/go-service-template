package main

import (
    "database/sql"
    "errors"
    "github.com/Gamma169/go-server-helpers/db"
    "github.com/google/uuid"
)


type SubModel struct {
    Id             string     `json:"id" jsonapi:"primary,sub-models"`
    FoobarModelId  *string    `json:"foobarModelId" jsonapi:"attr,foobar-model-id"`
    Value          string     `json:"value" jsonapi:"attr,value"`
    ValueInt       int        `json:"valueInt" jsonapi:"attr,value-int"`
    TempID         string     `json:"__id__" jsonapi:"attr,__id__"`
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
        foobarModelId   sql.NullString
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
            s.user_id, s.foobar_model, s.value, s.value_int
            FROM sub_models s
        WHERE s.user_id = ($1);
    `); err != nil {
        panic(err)
    }

}



/*********************************************
 * Database Functions
 * *******************************************/