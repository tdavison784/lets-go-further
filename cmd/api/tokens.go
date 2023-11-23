package main

import (
	"errors"
	"fmt"
	"greenlight.twd.net/internal/data"
	"greenlight.twd.net/internal/validator"
	"net/http"
	"time"
)

func (app *application) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userData, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.emailNotFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	token, err := app.models.Tokens.New(userData.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	tokenData := map[string]any{
		"activationToken": token.Plaintext,
		"userName":        userData.Name,
	}

	err = app.mailer.Send(input.Email, "new_token.tmpl", tokenData)
	if err != nil {
		app.logger.Error("Failed to send new activation token", "error", err)
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"token": fmt.Sprintf("Email with token activation details sent to %s", userData.Email)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// run validation steps to check that email and password meet defined criteria
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePassword(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Lookup user record by email provided in the client request
	// if no matching user was found, then we call the app.invalidCredentialResponse helper
	// to send a 401 unauthorized response to the client
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// check if the provided password matches the actual password for the user
	matches, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// if passwords don't match, we then call the invalidCredentialResponse helper
	if !matches {
		app.invalidCredentialResponse(w, r)
		return
	}

	// otherwise, if the password is correct we generate a new token with a 24 hour ttl
	// and scope 'authentication'
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
