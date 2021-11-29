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

// docker run -d --name=foobar_post -e POSTGRES_HOST_AUTH_METHOD=trust --net=host postgres
// docker exec -it foobar_post psql -h localhost -U postgres -p 5432 -c 'CREATE DATABASE foo;'

// DATABASE_NAME=foo DATABASE_USER=postgres DATABASE_HOST=127.0.0.1 ./bin/foobar

// Or
// ./scripts/setup_database


/*********************************************
 * Globals
 * *******************************************/

// An all-zero uuid constant.  Should not be considered valid by our system.
const ZERO_UUID = "00000000-0000-0000-0000-000000000000"

const SERVICE_PORT_ENV_VAR = "FOOBAR_PORT"
const DEFAULT_PORT = "7890"

var releaseMode string
var debug bool

var DB *sql.DB


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
    initMigrations()
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

    debugLog("foobar + DB connections Healthy (☆^ー^☆)")
    json.NewEncoder(w).Encode("foobar + DB connections Healthy! (☆^ー^☆)")
}


/*********************************************
 * Main
 * *******************************************/


func main() {
    router := mux.NewRouter()
    
    router.Path("/health").Methods("GET").HandlerFunc(HealthHandler)

    if getOptionalEnv("RUNNING_LOCALLY", "true") == "true" {
        AddCORSMiddlewareAndEndpoint(router)
    }

    router.Use(loggingMiddleware)




    // TODO
    s := router.PathPrefix("/user").Subrouter()
    s.Use(UserIdHeaderMiddleware)
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
    debugLog(fmt.Sprintf("Listening on port: %s\n", port))

    if debug {
        WalkRouter(router)
    }
    
    // Run our server in a goroutine so that it doesn't block.
    go func() {
        if err := server.ListenAndServe(); err != nil {
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
    shutdown()
    debugLog("Shutting down")
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