package main

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"math/rand"
	"time"
)

func initMigrations() {

	// If we have multiple services starting up at the same time
	// we don't want the migrations to overlap
	time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)

	debugLog("Doing Migrations")
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		logError(errors.New("Error:  Couldn't create migrations driver"), nil)
		logError(err, nil)
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations/",
		"postgres", driver)
	if err != nil {
		logError(errors.New("Error:  Couldn't run migrations"), nil)
		logError(err, nil)
		panic(err)
	}

	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			debugLog("no change")
		} else {
			logError(errors.New("Error with Migrations"), nil)
			panic(err)
		}
	}

	debugLog("Migrations Successful")
}
