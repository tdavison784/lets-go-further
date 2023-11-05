package main

import (
	"fmt"
	"net/http"
)

// delcare a handler which writes plain text response with all needed info back to end user
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	js := `{"Status": "Available", "Environment": %q, "Version": %q }`
	js = fmt.Sprintf(js, app.config.env, version)

	// Set content-type to application json on the response header
	w.Header().Set("Content-Type", "application/json")

	// write the JSON body as a response
	w.Write([]byte(js))
}
