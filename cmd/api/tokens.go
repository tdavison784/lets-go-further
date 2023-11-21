package main

import (
	"errors"
	"fmt"
	"greenlight.twd.net/internal/data"
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
