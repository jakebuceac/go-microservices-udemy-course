package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type RpcPayload struct {
	Name string
	Data string
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
		app.logItemViaRpc(response, requestPayload.Log)
	case "mail":
		app.sendMail(response, requestPayload.Mail)
	default:
		app.errorJSON(response, errors.New("unknown action"))
	}
}

func (app *Config) LogViaGrpc(response http.ResponseWriter, request *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(response, request, &requestPayload)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	connection, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	defer connection.Close()

	client := logs.NewLogServiceClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	_, err = client.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Logged"

	app.writeJSON(response, http.StatusAccepted, payload)
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

	request.Header.Set("Content-Type", "application/json")

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

func (app *Config) sendMail(response http.ResponseWriter, mailPayload MailPayload) {
	jsonData, _ := json.MarshalIndent(mailPayload, "", "\t")

	// Call the mail service
	mailServiceUrl := "http://mail-service/send"

	// Post to mail service
	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))

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

	// Check status code
	if clientResponse.StatusCode != http.StatusAccepted {
		app.errorJSON(response, errors.New("error calling mail service"))

		return
	}

	// Send back json
	var payload jsonResponse

	payload.Error = false
	payload.Message = "Message sent to: " + mailPayload.To

	app.writeJSON(response, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbit(response http.ResponseWriter, logPayload LogPayload) {
	err := app.pushToQueue(logPayload.Name, logPayload.Data)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Logged via RabbitMQ"

	app.writeJSON(response, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name string, message string) error {
	emitter, err := event.NewEventEmitter(app.RabbitMq)

	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	jsonData, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(jsonData), "log.INFO")

	if err != nil {
		return err
	}

	return nil
}

func (app *Config) logItemViaRpc(response http.ResponseWriter, logPayload LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	rpcPayload := RpcPayload{
		Name: logPayload.Name,
		Data: logPayload.Data,
	}

	var result string

	err = client.Call("RpcServer.LogInfo", rpcPayload, &result)

	if err != nil {
		app.errorJSON(response, err)

		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(response, http.StatusAccepted, payload)
}
