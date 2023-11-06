package main

import (
	"encoding/json"
	"greenlight.twd.net/internal/data"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the info we expect our clients to pass in the HTTP request body.
	// Note the field names and types in the struct are a subset of the Movie struct we created earlier.
	// This will be our *target decode destination*
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// init a new json.Decoder instance which rads from the request body and then uses the Decode() method to decode
	// the body contents into the input struct as the target decode destination. If there was an error, we use the
	// generic errorResponse() helper to send the client a 400 Bad Request response containing the error code
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Dump the contents of input struct into HTTP response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": input}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showMovieHandler accepts GET requests with URL params or JSON payload /v1/movies/:id
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "CasaBlanca",
		Runtime:   102,
		Genres:    []string{"romance", "drama", "war"},
		Year:      1982,
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
