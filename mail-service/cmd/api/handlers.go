package main

import "net/http"

func (app *Config) SendMail(response http.ResponseWriter, request *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	err := app.readJSON(response, request, &requestPayload)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	message := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSmtpMessage(message)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Sent to " + requestPayload.To,
	}

	app.writeJSON(response, http.StatusAccepted, payload)
}
