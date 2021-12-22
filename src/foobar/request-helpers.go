package main

import (
	"encoding/json"
    "errors"
    "github.com/google/jsonapi"
    "net/http"
)

const ContentTypeHeader = "Content-Type"
const AcceptContentTypeHeader = "Accept"
const JSONContentType = "application/json"

type InputObject interface {
	Validate() error
}

func PreProcessInput(input InputObject, r *http.Request) (error, int) {
	var err error

    header := r.Header.Get(ContentTypeHeader)
    if header == JSONContentType {
        
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


func sendErrorOnError(err error, status int, w http.ResponseWriter, r *http.Request) {
    if err != nil {
        logError(err, r)
        http.Error(w, err.Error(), status)
    }
}


func WriteModelToResponse(dataToSend interface{}, w *http.ResponseWriter, r *http.Request) error {
    
    debugLog(dataToSend)


    (*w).WriteHeader(http.StatusOK)
    header := r.Header.Get(ContentTypeHeader)
    acceptHeader := r.Header.Get(AcceptContentTypeHeader)
    if header == jsonapi.MediaType || acceptHeader == jsonapi.MediaType {
        (*w).Header().Set(ContentTypeHeader, jsonapi.MediaType)
        return jsonapi.MarshalPayload(*w, dataToSend)
    }
    
    (*w).Header().Set(ContentTypeHeader, JSONContentType)
    return json.NewEncoder(*w).Encode(dataToSend)
    return nil
}
