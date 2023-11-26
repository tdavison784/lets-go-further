package main

import (
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"greenlight.twd.net/internal/data"
	"greenlight.twd.net/internal/validator"
	"net"
	"net/http"
	"strings"
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
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
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

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to
		// any caches that the response may vary based on the value of the Authorization
		// header in the request.
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request.
		// This will return the empty string "" if there is no such header found
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization header found, use the contextSetUser()
		// helper that we just made to add the AnonymousUser to the request context.
		// Then we call the next handler in the chain and return without executing
		// any of the code below.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>". We try to split this into it's constituent parts, and if the
		// header isn't in the exact format we return a 401 Unauthorized response
		// using the invalidAuthenticationResponse helper
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// extract the actual token
		token := headerParts[1]

		// validator token to ensure it is in proper format
		v := validator.New()

		if data.ValidateToken(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// retrieve the details of the user associated with the token
		// again calling the invalidAuthenticationTokenResponse() helper
		// if no matching record was found.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// call the contextSetUser() helper to add the user information
		// to the request context
		r = app.contextSetUser(r, user)

		// call the next HTTP handler in the chain
		next.ServeHTTP(w, r)
	})
}

// requireActivatedUser checks that a user is both authenticated and activated
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// use the contextGetUser() helper that we made to retrieve the user information
		// from the request context
		user := app.contextGetUser(r)

		// if the user is not activated, then we call the inactiveAccountResponse helper
		// to inform the client that they must activate their account and try again
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		// call the next handler in the chain
		next.ServeHTTP(w, r)
	})
	// wrap the fn with the requireAuthenticatedUser middleware before returning it
	return app.requireAuthenticatedUser(fn)
}

// requireAuthenticatedUser checks that a user is not anonymous
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requirePermissions checks to see if the requesting user has the appropriate permissions to
// access the underlying resources. This wraps around the requireAuthenticatedUser and requireActivatedUser
// middleware to perform three checks at once
func (app *application) requirePermissions(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// retrieve user from context
		user := app.contextGetUser(r)

		// Get the slice of permissions for the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		app.logger.Info("User permissions", "permissions", permissions)

		// check if the slice includes required permissions. If it doesn't then we return
		// 403 forbidden response
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// otherwise they have the required permissions so we call the next handler in the chain
		next.ServeHTTP(w, r)
	}
	return app.requireActivatedUser(fn)
}
