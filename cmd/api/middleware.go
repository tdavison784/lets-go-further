package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
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

// rateLimit method puts I.P based rate limiting to work by creating a custom
// map to hold Client I.P information as well as a rate.Limiter
func (app *application) rateLimit(next http.Handler) http.Handler {
	// client struct used to hold rate limiter and last seen time
	// for each client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the client's I.P addresses and rate limiters
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map
	// once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three minutes,
			// delete the corresponding entry from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Importantly, unlock the mutex when the cleanup is complete
			mu.Unlock()
		}
	}()
	// the function we are returning is a closure, which 'closes over' the limiter
	// variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Extract the client's I.P address from the request
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// lock the mutex to prevent this code from being executed concurrently
		mu.Lock()

		// check to see if the I.P address already exists in the map. If it doesn't
		// then initialize a new rate limiter and add the I.P address and limiter to the map
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}

		// update the clients lastSeen time
		clients[ip].lastSeen = time.Now()
		// Call limiter.Allow() to see if te request is permitted, and if it's not
		// then we call the rateLimitExceededResponse() helper to return a 429 Too Many Requests
		// HTTP Response code.
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		// VERY IMPORTANT, unlock the mutex before calling the next handler in the chain.
		// Notice that we DON'T use defer to unlock the mutex, as that would mean that
		// the mutex isn't unlocked until all the handlers downstream of this middleware
		// have also returned
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
