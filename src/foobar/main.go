package main

import (
	"database/sql"
	"encoding/json"
	"errors"
  "github.com/Gamma169/go-server-helpers/db"
  envs "github.com/Gamma169/go-server-helpers/environments"
	"github.com/Gamma169/go-server-helpers/server"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"time"
)

// docker run -d --name=foobar_post -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=foo -p 5432:5432 postgres:9.6.17-alpine

// Local
// ./scripts/build.sh foobar && DATABASE_NAME=foo DATABASE_USER=postgres DATABASE_HOST=127.0.0.1 RUN_MIGRATIONS=true ./bin/foobar

// Docker
// docker run -it --rm -e DATABASE_NAME=foo  -e DATABASE_HOST=127.0.0.1 -e DATABASE_USER=postgres -e RUN_MIGRATIONS=true --net=host  --name=foobar gamma169/foobar

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
	if releaseMode = envs.GetOptionalEnv("RELEASE_MODE", "dev"); releaseMode != "production" {
		debug = true
	}

	// Note that we check any required environment variables before anything else
	// in order not to create any "hanging" db connections and immediately terminate
	// if we are missing any down the line
	checkRequiredEnvs()
	DB = db.InitDB("", debug)

	if envs.GetOptionalEnv("RUN_MIGRATIONS", "false") == "true" {
		initMigrations()
	}

	initFoobarModelsPreparedStatements()
}

func checkRequiredEnvs() {
	// Not sure if I should use the getOptionalEnv function here or just os.LookupEnv
	// Because if I use getOptionalEnv and it doesn't exist, we output the logs for it twice
	// I think that's fine, but I need to think on it
	if envs.GetOptionalEnv("DATABASE_URL", "") == "" {
		envs.GetRequiredEnv("DATABASE_NAME")
		envs.GetRequiredEnv("DATABASE_HOST")
		envs.GetRequiredEnv("DATABASE_USER")
	}
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

	if envs.GetOptionalEnv("RUNNING_LOCALLY", "true") == "true" {
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


	port := envs.GetOptionalEnv(SERVICE_PORT_ENV_VAR, envs.GetOptionalEnv("PORT", DEFAULT_PORT))
	server.SetupAndRunServer(router, port, debug, shutdown)
}
