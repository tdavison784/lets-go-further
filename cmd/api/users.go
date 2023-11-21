package main

import (
	"errors"
	"greenlight.twd.net/internal/data"
	"greenlight.twd.net/internal/validator"
	"net/http"
	"time"
)

// registerUserHandler is used to create new users in our system
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	// create anonymous struct to hold client data
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// copy the data from the input struct into a new User struct.
	// Notice also that we set the Activated field to false, which
	// isn't strictly necessary because the Activated field will have
	// the zero-value of false by default. But setting this explicitly
	// helps to make our intentions clear to anyone reading the code
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Use the Password.Set() method to generate and store the hashed
	// and plaintext passwords.
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	// validate the user struct and return the error message to the
	// client if any validation checks fail
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert new user into database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get a ErrDuplicateEmail error, use the v.AddError() method to manually
		// add a message to the validator instance, and then call our failedValidationResponse() helper
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// After the user record has been created in the database,
	// generate a new activation token for the user
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Call the Send() method on our Mailer, Passing in the user's email address
	// name of the template file, and the User struct containing the new users data
	// below we place the email confirmation sending into a custom helper
	// to run the code in a goroutine in the background
	// this improves the HTTP request time significantly as we are no longer waiting for the
	// email to be sent before returning a response to our End user.

	app.background(func() {

		// As there are now multiple pieces of data that we want to pass to our email templates,
		// we create a map to act as a 'holding structure' for the data. This contains
		// the plaintext version of the activation token for the user, along with their User ID
		tokenData := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          token.UserID,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", tokenData)
		if err != nil {
			app.logger.Error("Failed to send confirmation email.", "error", err.Error())
		}
	})

	// write JSON response containing the data along with a 201 created status code
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// activateUserHandler is used to activate users based on their activation token
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateToken(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve the details of the user associated with the token using the
	// GetForToken() method. If no matching record is found, then we let
	// the client know that the token provided is not valid.
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// activate the user if all the above checks pass
	user.Activated = true

	// save the updated user record in our database, checking for any edit conflicts
	// in the same way we did for our movie records
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// if everything went successfully, then we delete all activation token for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// send updated user details to the client
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
