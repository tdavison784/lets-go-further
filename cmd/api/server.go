package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (app *application) server() error {
	// declare an HTTP server which listens on the port provided in the config struct as well as shows the env
	// we will use the servermux from above as the handler, give some timeout settings and add our structured logging
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// start the HTTP server
	app.logger.Info("Starting Server", "addr", srv.Addr, "env", app.config.env)

	return srv.ListenAndServe()
}
