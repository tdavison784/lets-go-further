package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"greenlight.twd.net/internal/data"
	"log/slog"
	"net/http"
	"os"
	"time"

	// import the pq driver so that it can register itself with the database/sql package.
	// Note that we alias this import to the blank identifier, to stop the Go compiler complaining
	// that the package is not used
	_ "github.com/lib/pq"
)

// declare a version of the api that we can return in the healthcheck
const version = "1.0.0"

// declare a struct that will hold configs for our HTTP server
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	// adding in a new limiter struct containing fields for the requests per second (rps)
	// and burst values, and a boolean field which we use to enable/disable rate limiting
	// altogether
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

// declare a struct that will hold all dependencies for our application's HTTP handlers, helpers, and middleware.
type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {

	// declare an instance of our config struct
	var cfg config

	//read the values of port and env from CLI flags
	flag.StringVar(&cfg.env, "env", "development", "Environment(DEV|QA|STAGE|PROD)")
	flag.IntVar(&cfg.port, "port", 4000, "HTTP Server Listening Port")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("postgresConnString"), "Postgres SQL connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "number of maximum open DB connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "number of maximum idle DB connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgresSQL max connection idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiting")
	flag.Parse()

	// initialize our logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// call openDB() helper function to establish a DB connection pool
	// we pass in our cfg struct, if this returns an error we log it and exit
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// set max open connections for the db sessions
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// set max connection idle time
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	// set max idle open connections
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// defer a call to db.Close() so that the connection pool is closed before main func exits
	defer db.Close()

	// log that a connection pool has been established
	logger.Info("Successfully established database connection")

	// declare an instance of the app struct, containing the config and our logger
	app := application{
		cfg,
		logger,
		data.NewModels(db),
	}

	// declare an HTTP server which listens on the port provided in the config struct as well as shows the env
	// we will use the servermux from above as the handler, give some timeout settings and add our structured logging
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// start the HTTP server
	logger.Info("Starting Server", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB() helper function returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	// use sql.Open() to create an empty connection pool, using the DSN from our
	// config struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// create a context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	// Use PingContext() to establish a new connection to the database
	// passing in the context we created above. If the connection could not be established
	// within the 5 second deadline, this will return an error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// otherwise return the db context and nil error
	return db, nil

}
