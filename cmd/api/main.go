package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// declare a version of the api that we can return in the healthcheck
const version = "1.0.0"

// declare a struct that will hold configs for our HTTP server
type config struct {
	port int
	env  string
}

// declare a struct that will hold all dependencies for our application's HTTP handlers, helpers, and middleware.
type application struct {
	config config
	logger *slog.Logger
}

func main() {

	// declare an instance of our config struct
	var cfg config

	//read the values of port and env from CLI flags
	flag.StringVar(&cfg.env, "env", "development", "Environment(DEV|QA|STAGE|PROD)")
	flag.IntVar(&cfg.port, "port", 4000, "HTTP Server Listening Port")
	flag.Parse()

	// initialize our logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// declare an instance of the app struct, containing the config and our logger
	app := application{
		cfg,
		logger,
	}

	// declare a new servermux and add the /v1/healthcheck route which dispatches requests to the
	// healthcheckHandler method
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// declare an HTTP server which listens on the port provided in the config struct as well as shows the env
	// we will use the servermux from above as the handler, give some timeout settings and add our structured logging
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// start the HTTP server
	logger.Info("Starting Server", "addr", srv.Addr, "env", cfg.env)

	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
