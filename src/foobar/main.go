package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
    "github.com/Gamma169/go-server-helpers/db"
    envs "github.com/Gamma169/go-server-helpers/environments"
	"github.com/Gamma169/go-server-helpers/server"
    "github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"time"
)

// docker run -d --name=foobar_post -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=foo -p 5432:5432 postgres:9.6.17-alpine
// docker run -d --name=foobar_redis -p 6379:6379 redis:6-alpine

// Local
// ./scripts/build.sh foobar && REDIS_HOST=127.0.0.1 DATABASE_NAME=foo DATABASE_USER=postgres DATABASE_HOST=127.0.0.1 RUN_MIGRATIONS=true ./bin/foobar

// Docker
// docker run -it --rm -e REDIS_HOST=127.0.0.1 -e DATABASE_NAME=foo  -e DATABASE_HOST=127.0.0.1 -e DATABASE_USER=postgres -e RUN_MIGRATIONS=true --net=host  --name=foobar gamma169/foobar

// Or
// ./scripts/setup_database.sh

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
var isRunningLocally bool

var DB *sql.DB
var redisClient *redis.Client

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
	if releaseMode = envs.GetOptionalEnv("RELEASE_MODE", "dev"); releaseMode != "production" {
		debug = true
	}
    if envs.GetOptionalEnv("RUNNING_LOCALLY", "true") == "true"{
        isRunningLocally = true
    }

	// Note that we check any required environment variables before anything else
	// in order not to create any "hanging" db connections and immediately terminate
	// if we are missing any down the line
	checkRequiredEnvs()
    
    redisClient = db.InitRedis("", false, debug)
    DB = db.InitPostgres("", debug)

	if envs.GetOptionalEnv("RUN_MIGRATIONS", "false") == "true" {
		db.InitPostgresMigrations(DB, 7000, isRunningLocally, debug)
	}

	initFoobarModelsPreparedStatements()
}

func checkRequiredEnvs() {
	db.CheckRequiredPostgresEnvs("")
	db.CheckRequiredRedisEnvs("", false)
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

    if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
        logError(errors.New("Error: Could not connect to Redis"), nil)
        panic(err)
    }

    debugLog(BluePrint + "foobar service + DB + Redis connections Healthy (☆^ー^☆)" + EndPrint)
    json.NewEncoder(w).Encode("foobar service + DB + Redis connections Healthy! (☆^ー^☆)")
}

/*********************************************
 * Main
 * *******************************************/

func main() {

	router := mux.NewRouter()
    server.AddLoggingMiddleware(router, TRACE_ID_HEADER, debug)

	router.Path("/health").Methods(http.MethodGet).HandlerFunc(HealthHandler)

	if isRunningLocally {
		server.AddCORSMiddlewareAndEndpoint(router, REQUESTER_ID_HEADER)
	}

	s := router.PathPrefix("/user").Subrouter()
	server.AddRequesterIdHeaderMiddleware(s, REQUESTER_ID_HEADER, debug)
	s.Path("/foobar-models").Methods(http.MethodGet).HandlerFunc(GetFoobarModelHandler)
	s.Path("/foobar-models").Methods(http.MethodPost).HandlerFunc(CreateOrUpdateFoobarModelHandler)
	s.Path("/foobar-models/{modelId}").Methods(http.MethodPatch).HandlerFunc(CreateOrUpdateFoobarModelHandler)
	s.Path("/foobar-models/{modelId}").Methods(http.MethodDelete).HandlerFunc(DeleteFoobarModelHandler)

	// s.Path("/").Methods("GET").HandlerFunc(GetUserHandler)

	// TODO
	// This route should always be at the bottom
	// router.Path("/{service:[a-zA-Z0-9_-]+}{endpoint:.*}").HandlerFunc(ProxyHandler)


	port := envs.GetOptionalEnv(SERVICE_PORT_ENV_VAR, envs.GetOptionalEnv("PORT", DEFAULT_PORT))
	server.SetupAndRunServer(router, port, debug, shutdown)
}
