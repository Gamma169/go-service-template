package main

import (
    "context"
    "encoding/json"
    "errors"
    "database/sql"
    "fmt"
    "github.com/gorilla/mux"
    "log"
    "math/rand"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "time"
)

// docker run -d --name=foobar_post -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=foo -p 5432:5432 postgres:9.6.17-alpine

// Local
// ./scripts/build foobar && DATABASE_NAME=foo DATABASE_USER=postgres DATABASE_HOST=127.0.0.1 RUN_MIGRATIONS=true ./bin/foobar

// Docker
// docker run -it --rm -e DATABASE_NAME=foo  -e DATABASE_HOST=127.0.0.1 -e DATABASE_USER=postgres -e RUN_MIGRATIONS=true --net=host  --name=foobar gamma169/foobar


// Or
// ./scripts/setup_database


/*********************************************
 * Globals Vars + Config Consts -- Some helper consts are in their respective files
 * *******************************************/

// An all-zero uuid constant.  Should not be considered valid by our system.
const ZERO_UUID = "00000000-0000-0000-0000-000000000000"
const TRACE_ID_HEADER = "request-id"
const REQUESTER_ID_HEADER = "user-id"

const SERVICE_PORT_ENV_VAR = "FOOBAR_PORT"
const DEFAULT_PORT = "7890"

const DB_ARRAY_DELIMITER = ":::"

var releaseMode string
var debug bool

var DB *sql.DB


var getFoobarModelsStmt *sql.Stmt
var postFoobarModelStmt *sql.Stmt
var updateFoobarStmt *sql.Stmt
var deleteFoobarStmt *sql.Stmt

/*********************************************
 * Init and Shutdown
 * *******************************************/

// This function is builtin go func that gets automatically called
func init() {
    rand.Seed(time.Now().UnixNano())
    // Note that we check any required environment variables before anything else
    // in order not to create any "hanging" db connections and immediately terminate
    // if we are missing any down the line
    checkRequiredEnvs()
    initDebug()
    initDB()
    
    if getOptionalEnv("RUN_MIGRATIONS", "false") == "true" {
        initMigrations()
    }

    initFoobarModelsPreparedStatements()
}


// This is a custom function that is called for graceful shutdown
func shutdown() {
    // close redis connections
    // redisClient.Close()
}


/*********************************************
 * Health Handler
 * *******************************************/


func HealthHandler(w http.ResponseWriter, r *http.Request) {

    if err := DB.Ping(); err != nil {
        logError(errors.New("Error: Could not connect to DB"), r)
        panic(err)
    }

    debugLog(BluePrint + "foobar + DB connections Healthy (☆^ー^☆)" + EndPrint)
    json.NewEncoder(w).Encode("foobar + DB connections Healthy! (☆^ー^☆)")
}


/*********************************************
 * Main
 * *******************************************/


func main() {
    router := mux.NewRouter()
    router.Use(loggingMiddleware)
    
    router.Path("/health").Methods(http.MethodGet).HandlerFunc(HealthHandler)

    if getOptionalEnv("RUNNING_LOCALLY", "true") == "true" {
        AddCORSMiddlewareAndEndpoint(router)
    }



    s := router.PathPrefix("/user").Subrouter()
    s.Use(RequesterIdHeaderMiddleware)
    s.Path("/foobar-models").Methods(http.MethodGet).HandlerFunc(GetFoobarModelHandler)
    s.Path("/foobar-models").Methods(http.MethodPost).HandlerFunc(CreateOrUpdateFoobarModelHandler)
    s.Path("/foobar-models/{modelId}").Methods(http.MethodPatch).HandlerFunc(CreateOrUpdateFoobarModelHandler)
    s.Path("/foobar-models/{modelId}").Methods(http.MethodDelete).HandlerFunc(DeleteFoobarModelHandler)
    
    // s.Path("/").Methods("GET").HandlerFunc(GetUserHandler)

    
    // TODO
    // This route should always be at the bottom
    // router.Path("/{service:[a-zA-Z0-9_-]+}{endpoint:.*}").HandlerFunc(ProxyHandler)


    port := getOptionalEnv(SERVICE_PORT_ENV_VAR, getOptionalEnv("PORT", DEFAULT_PORT))
    server := http.Server{
      Handler: router,
      Addr: fmt.Sprintf("0.0.0.0:%s", port),
      WriteTimeout: 5 * time.Minute,
      ReadTimeout:  5 * time.Minute,
    }
    // This should be the only 'log' so that we have at least one line printed when the server starts in production mode
    log.Println("Server started -- Ready to accept connections")
    debugLog(fmt.Sprintf("Listening on port: %s", port))

    if debug {
        WalkRouter(router)
    }
    
    // Run our server in a goroutine so that it doesn't block.
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logError(err, nil)
        }
    }()


    // Graceful shutdown procedure taken from example:
    // https://github.com/gorilla/mux#graceful-shutdown
    c := make(chan os.Signal, 1)
    // We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
    // SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
    signal.Notify(c, os.Interrupt)
    // Block until we receive our signal.
    <-c
    // Create a deadline to wait for.
    var waitTime = time.Second * 15
    if debug {
        waitTime = time.Millisecond * 500
    }
    ctx, cancel := context.WithTimeout(context.Background(), waitTime)
    defer cancel()
    // Doesn't block if no connections, but will otherwise wait
    // until the timeout deadline.
    server.Shutdown(ctx)
    debugLog("Shutting down")
    shutdown()
    log.Println("Completed shutdown sequence.  Thank you and goodnight.  <(_ _)>")
    os.Exit(0)
}


// Print out all route info
// Walk function taken from example
// https://github.com/gorilla/mux#walking-routes
func WalkRouter(router *mux.Router) {
    err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
        pathTemplate, err := route.GetPathTemplate()
        if err == nil {
            fmt.Println("ROUTE:", pathTemplate)
        }
        pathRegexp, err := route.GetPathRegexp()
        if err == nil {
            fmt.Println("Path regexp:", pathRegexp)
        }
        queriesTemplates, err := route.GetQueriesTemplates()
        if err == nil {
            fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
        }
        queriesRegexps, err := route.GetQueriesRegexp()
        if err == nil {
            fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
        }
        methods, err := route.GetMethods()
        if err == nil {
            fmt.Println("Methods:", strings.Join(methods, ","))
        }
        fmt.Println()
        return nil
    })
    if err != nil {
        fmt.Println(err)
    }
}