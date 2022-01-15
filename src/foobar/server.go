package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)


func SetupAndRunServer(router *mux.Router) {
	port := getOptionalEnv(SERVICE_PORT_ENV_VAR, getOptionalEnv("PORT", DEFAULT_PORT))
	server := http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
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
