package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(response http.ResponseWriter, request *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "You have hit the broker",
	}

	_ = app.writeJSON(response, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(response http.ResponseWriter, request *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(response, request, &requestPayload)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(response, requestPayload.Auth)
	case "log":
		app.logItem(response, requestPayload.Log)
	default:
		app.errorJSON(response, errors.New("unknown action"))
	}
}

func (app *Config) logItem(response http.ResponseWriter, logPayload LogPayload) {
	jsonData, _ := json.MarshalIndent(logPayload, "", "\t")
	logServiceUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	clientResponse, err := client.Do(request)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	defer clientResponse.Body.Close()

	if clientResponse.StatusCode != http.StatusAccepted {
		app.errorJSON(response, err)

		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Logged!"

	app.writeJSON(response, http.StatusAccepted, payload)
}

func (app *Config) authenticate(response http.ResponseWriter, authPayload AuthPayload) {
	// Create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(authPayload, "", "\t")

	// Call service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	client := &http.Client{}
	clientResponse, err := client.Do(request)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	defer clientResponse.Body.Close()

	// Make sure we get back the correct status code
	if clientResponse.StatusCode == http.StatusUnauthorized {
		app.errorJSON(response, errors.New("invalid credentials"))

		return
	} else if clientResponse.StatusCode != http.StatusAccepted {
		app.errorJSON(response, errors.New("error calling authentication service"))

		return
	}

	// Create a variable we'll read clientResponse.Body into
	var jsonFromService jsonResponse

	// Decode the json from the authenction service
	err = json.NewDecoder(clientResponse.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	if jsonFromService.Error {
		app.errorJSON(response, err, http.StatusUnauthorized)

		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJSON(response, http.StatusAccepted, payload)
}
