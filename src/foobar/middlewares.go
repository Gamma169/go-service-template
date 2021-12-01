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
    router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func (w http.ResponseWriter, r *http.Request) {w.WriteHeader(http.StatusNoContent)})
}


// Check if the user-id header exists and return 400 if not
func UserIdHeaderMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        
        requesterId := r.Header.Get(REQUESTER_ID_HEADER)
        if requesterId == "" {
            msg := "No '" + REQUESTER_ID_HEADER + "' header"
            debugLog(msg)
            http.Error(w, msg, http.StatusBadRequest)
            return
        }
        if _, err := uuid.Parse(requesterId); err != nil {
            msg := REQUESTER_ID_HEADER + "- is not valid UUID"
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
        requestId := r.Header.Get(TRACE_ID_HEADER)
        if requestId == "" {
            requestId = uuid.New().String()
            r.Header.Set(TRACE_ID_HEADER, requestId)
        }

        // I thought of using fmt.Sprintf, but it seems that the plus sign is actually the most efficient
        // TODO: Check/benchmark efficiency of this specific use case
        debugLog(BoldPrint, HeaderPrint, "Recieved:", r.RequestURI, "--", requestId,  EndPrint)
        next.ServeHTTP(w, r)
        // TODO: Prob implement custom responseWriter to be able to get status from request
        // https://github.com/Gamma169/go-service-template/issues/2
        debugLog(BoldPrint, HeaderPrint, "Finished:", r.RequestURI, "--", requestId, "--", "[]", EndPrint)
    })
}
