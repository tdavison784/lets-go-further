package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic as Go unwinds the stack
		defer func() {
			// Use the builtin recover function to check if there has been a panic or not
			if err := recover(); err != nil {
				// if there was a panic, set a "Connection: close" header on the response.
				// This acts as a trigger to make Go's HTTP server automatically close the current connection after
				// a response has been sent.
				w.Header().Set("Connection", "close")

				// The value returend by recover() has the type any, so we can use fmt.Errorf() to normalize it
				// into an error and call our serverErrorResponse Helper method. In turn, this will log the error
				// using our custom logger type at the Error level and send the client a 500 internal serv resp error
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Initialize a new rate limiter which allows an average of 2 request a second
	// and a max of 4 requests in a single 'burst'
	limiter := rate.NewLimiter(2, 4)

	// the function we are returning is a closure, which 'closes over' the limiter
	// variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call limiter.Allow() to see if te request is permitted, and if it's not
		// then we call the rateLimitExceededResponse() helper to return a 429 Too Many Requests
		// HTTP Response code.
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
