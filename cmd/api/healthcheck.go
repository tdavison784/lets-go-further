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

	// uncomment this as a quick way to test graceful shutdowns
	// on Mac use ps -ef |grep api command to get the P.I.D
	// then run curl localhost:<port>/v1/healthcheck & pkill -15 -P <P.I.D>
	// in a separate terminal window. You will see the app server log the shutdown started
	// and shutdown complete messages after the requests have finished.
	//time.Sleep(4*time.Second)

	err := app.writeJSON(w, http.StatusOK, envelope{"status": data}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
