package main

import (
	"greenlight.twd.net/internal/data"
	"greenlight.twd.net/internal/validator"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the info we expect our clients to pass in the HTTP request body.
	// Note the field names and types in the struct are a subset of the Movie struct we created earlier.
	// This will be our *target decode destination*
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Init new Validator instance
	v := validator.New()

	// Use the Check method to execute validation checks. This will add the provided key and error message
	// to the errors map if the check does not evaluate to true.
	v.Check(input.Title != "", "title", "title must not be empty")
	v.Check(len(input.Title) <= 500, "title", "title must be less then 500 bytes long")

	v.Check(input.Year != 0, "year", "must provide a year")
	v.Check(input.Year >= 1888, "year", "year must be greater then 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "year cannot be in the future")

	v.Check(input.Genres != nil, "genres", "genres must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(input.Genres) <= 5, "genres", "cannot contain more then 5 genres")
	v.Check(validator.Unique(input.Genres), "genres", "cannot contain duplicate genres")

	// Use the Valid() method to see if any checks failed. If they did, then use the failedValidationResponse
	// helper to send a response to the client, passing in the v.Errors map
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
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
