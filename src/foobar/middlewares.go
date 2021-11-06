package main

import (
    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "net/http"
)


func AddCORSMiddlewareAndEndpoint(router *mux.Router) {
    // This adds CORS headers for all requests if running locally
    router.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Credentials", "true")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Session")
            w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
            next.ServeHTTP(w, r)
        })
    })
    // Any options requests return 204-- to be used with above CORS stuff
    router.PathPrefix("/").Methods("OPTIONS").HandlerFunc(func (w http.ResponseWriter, r *http.Request) {w.WriteHeader(http.StatusNoContent)})
}


// Check if the user-id header exists and return 400 if not
func UserIdHeaderMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        
        userId := r.Header.Get("user-id")
        
        if userId == "" {
            msg := "No 'user-id' header"
            debugLog(msg)
            http.Error(w, msg, http.StatusBadRequest)
            return
        }
        if _, err := uuid.Parse(userId); err != nil {
            msg := "'user-id' is not valid UUID"
            debugLog(msg)
            http.Error(w, msg, http.StatusBadRequest)
            return   
        }

        // Call the next handler, which can be another middleware in the chain, or the final handler.
        next.ServeHTTP(w, r)
    })
}



func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        debugLog(r.RequestURI)
        next.ServeHTTP(w, r)
    })
}
