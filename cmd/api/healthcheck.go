package main

import (
	"net/http"
)

// delcare a handler which writes plain text response with all needed info back to end user
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	// create a map that has all healthcheck info
	data := map[string]string{
		"Status":      "Available",
		"Environment": app.config.env,
		"Version":     version,
	}

	err := app.writeJSON(w, 200, envelope{"status": data}, nil)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The application encountered a problem and could not complete your request",
			http.StatusInternalServerError)
	}
}
