package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Create a new movie")
}

// showMovieHandler accepts GET requests with URL params or JSON payload /v1/movies/:id
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	// when httprouter is parsing a request, any interpolated URL params will be stored in the request context.
	// We can use the ParamsFromContext() function to retrieve a slice containing these parameter names and values.
	params := httprouter.ParamsFromContext(r.Context())

	// We can then use the ByName() method to get the value of the "id" parameter which will always be a positive int.
	// ByName() returns only string values so we need to convert it to an int with a base of 10, and 64 bits.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// otherwise return the value of the id lookup
	fmt.Fprintf(w, "show the details of movie %d", id)

}
