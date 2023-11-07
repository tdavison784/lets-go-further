package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	// Decode the request body into the target destination
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		// if there is an error during decoding, start the triage
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// use the errors.As() method to check whether the error has the type *json.SyntaxError
		// if it does, then return a plain-english error message which includes the location of the problem
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON. So we check for this using errors.Is() and
		// return a generic error message.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
		// JSON value is the wrong type for the target destination. If the error relates
		// to a specific field, then we include that in our error message to make it easier
		// for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)",
				unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty.
		// We check for this with errors.Is() and return a plain-english error message instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// A json.InvalidUnmarshalError error will be returned if we pass something that is a non-nil pointer
		// to decode(). We catch this and panic, rather than returning an error to our handler. At the end of
		// this chapter we'll talk about panicking versus returning errors, and discuss why it's an appropriate
		// thing to do in this specific situation.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format: "json:unknown field "<name>""
		// We check for this, extract the field name from the error, and interpolate it into our custom
		// error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains an unknown key %s", fieldName)

		// Use the errors.As() method to check whether the error has the type *http.MaxBytesError.
		// If it does, then it means the request body exceeded our size limit of 1MB, and we return a clear message
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		// for everything else return an error message as is
		default:
			return err
		}
	}
	// Call Decode() again, using a pointer to an empty anonymous struct as the dest
	// if the request body only contained a single JSON value this will return an io.EOF error.
	// so if we get anything else, we know that there is an additional data in the request body, and we return our own
	// custom error message.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
