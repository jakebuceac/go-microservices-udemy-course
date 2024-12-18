package main

import (
	"log-service/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(response http.ResponseWriter, request *http.Request) {
	// Read json into var
	var requestPayload JSONPayload

	_ = app.readJSON(response, request, &requestPayload)

	// Insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Logged",
	}

	app.writeJSON(response, http.StatusAccepted, payload)
}
