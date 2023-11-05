package main

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

// Retrieve the "id" URL parameter from the current request context, convert it to an int and return it.
// if the operation isn't successful return 0 and an error
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// writeJSON() helper for sending responses. This takes the destination http.ResonseWriter, the HTTP status code to send
// the data to encode into JSON and a header map containing any additional HTTP headers we want to include in the resp
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {

	//encode the data param to JSON, return err if there were any
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	// append a newline for read-ability
	js = append(js, '\n')

	// At this point we know the JSON struct is fine so we can write any additonal headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header then write the status code and JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// Define an envelope type
type envelope map[string]any
