package main

import (
	"errors"
	"greenlight.twd.net/internal/data"
	"greenlight.twd.net/internal/validator"
	"net/http"
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

	// Call the Send() method on our Mailer, Passing in the user's email address
	// name of the template file, and the User struct containing the new users data
	err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// write JSON response containing the data along with a 201 created status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
