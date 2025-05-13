package main

import (
	"bytes"
	"encoding/json"
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
	user, err := app.Repo.GetByEmail(requestPayload.Email)

	if err != nil {
		app.errorJSON(response, errors.New("invalid credentials"), http.StatusBadRequest)

		return
	}

	valid, err := app.Repo.PasswordMatches(requestPayload.Password, *user)

	if err != nil || !valid {
		app.errorJSON(response, errors.New("invalid credentials"), http.StatusBadRequest)

		return
	}

	// Log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in!", user.Email))

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(response, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name string, data string) error {
	var logEntry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	logEntry.Name = name
	logEntry.Data = data

	jsonData, _ := json.MarshalIndent(logEntry, "", "\t")
	logServiceUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		return err
	}

	_, err = app.Client.Do(request)

	if err != nil {
		return err
	}

	return nil
}
