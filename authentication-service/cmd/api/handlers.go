package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) Authenticate(response http.ResponseWriter, request *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(response, request, &requestPayload)

	if err != nil {
		app.errorJSON(response, err, http.StatusBadRequest)

		return
	}

	// Validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)

	if err != nil {
		app.errorJSON(response, errors.New("invalid credentials"), http.StatusBadRequest)

		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)

	if err != nil || !valid {
		app.errorJSON(response, errors.New("invalid credentials"), http.StatusBadRequest)

		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(response, http.StatusAccepted, payload)
}
