package main

import (
	"encoding/json"
    "errors"
    "github.com/google/jsonapi"
    "net/http"
)

type InputObject interface {
	Validate() error
}

func PreProcessInput(input InputObject, r *http.Request) (error, int) {
	var err error

    header := r.Header.Get("Content-Type")
    
    if header == "application/json" {
        
        dec := json.NewDecoder(r.Body)
        dec.DisallowUnknownFields()

        if err = dec.Decode(input); err != nil {
            return err, http.StatusBadRequest
        }

    } else if header == jsonapi.MediaType {
        
        if err := jsonapi.UnmarshalPayload(r.Body, &input); err != nil {
            return err, http.StatusBadRequest
        }

    } else {
        return errors.New("Content-Type header is not json or jsonapi standard"), http.StatusBadRequest
    }


    if err = input.Validate(); err != nil {
        return err, http.StatusBadRequest
    }

    return nil, 0
}
