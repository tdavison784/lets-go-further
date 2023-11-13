package main

import (
	"errors"
	"fmt"
	"greenlight.twd.net/internal/data"
	"greenlight.twd.net/internal/validator"
	"net/http"
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

	// Copy values from input struct into Movie struct
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Init new Validator instance
	v := validator.New()

	// Call the ValidateMovie() method and return a response containing the errors if any of the checks fail
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// when sending a HTTP response, we want to include a Location Header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location Header.
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Dump the contents of input struct into HTTP response
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
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

	// call Get() method to fetch a record from the DB by its ID.
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound error
	// in which case we send a 404 Not found response to the client
	movie, err := app.models.Movies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateMovieHandler accepts PUT requests
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// call Get() method to fetch a record from the DB by its ID.
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound error
	// in which case we send a 404 Not found response to the client
	movie, err := app.models.Movies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// declare an input struct to hold the expected data from the client
	// in order to be able to do partial updates against the below struct
	// we are going to use pointers to the underlying types
	// we do this because when using a pointer to the type we can check to see
	// if a user supplied a value for it, if they did the value will not be nil
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// we check the struct to see if the pointer values are equal to their own types nil equivalent
	// we are doing this, so we can handle partial updates
	// If the fields we check are not equal to their nil equivalents
	// copy the values from the request body to the appropriate fields of the movie record.
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// run the validation checks
	// Init new Validator instance
	v := validator.New()

	// Call the ValidateMovie() method and return a response containing the errors if any of the checks fail
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// deleteMovieHandler accepts DELETE request
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete movie from DB and return a 404 error if the DB record is not found
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMovieHandler(w http.ResponseWriter, r *http.Request) {
	// to keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	// create a new validator instance
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data
	qs := r.URL.Query()

	// Use our helpers to extract the title and genres string values, failing back on defaults
	// of an empty string and empty slice if they were not provided by the client
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// get the page and page_size query string values as integers.
	// notice we set the default page value to 1 and page_size to 20
	// and that we pass the validator instance as the final argument here
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply a ascending sort on movie ID)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// Execute validation checks on the Filters struct and send a response containing the errors
	// if necessary
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to retrieve the movies, passing in the various filter parameters
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
