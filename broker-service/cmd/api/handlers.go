package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zahariaca/broker/event"
	"github.com/zahariaca/broker/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"net/rpc"
	"time"

	"github.com/zahariaca/toolbox"
)

var tools toolbox.Tools

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

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := toolbox.JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = tools.WriteJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := tools.ReadJson(w, r, &requestPayload)
	if err != nil {
		tools.ErrorJson(w, err)
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItem(w, requestPayload.Log)
	case "logViaRabbitMQ":
		app.logEventViaRabbit(w, requestPayload.Log)
	case "logViaRPC":
		app.logViaRPC(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)

	default:
		tools.ErrorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code

	if response.StatusCode == http.StatusUnauthorized {
		tools.ErrorJson(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		tools.ErrorJson(w, errors.New("error calling auth service"))
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService toolbox.JsonResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	if jsonFromService.Error {
		tools.ErrorJson(w, err, http.StatusUnauthorized)
		return
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: "Authenticated",
		Data:    jsonFromService.Data,
	}

	tools.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		tools.ErrorJson(w, err)
		return
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: "logged",
	}

	tools.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, mail MailPayload) {
	jsonData, _ := json.MarshalIndent(mail, "", "\t")

	mailerServiceURL := "http://mailer-service/send"

	request, err := http.NewRequest("POST", mailerServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		tools.ErrorJson(w, err)
		return
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Message sent to: %s", mail.To),
	}

	tools.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, entry LogPayload) {
	err := app.pushToQueue(entry.Name, entry.Data)

	if err != nil {
		tools.ErrorJson(w, err)
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: "logged via RabbitMQ",
	}

	err = tools.WriteJson(w, http.StatusAccepted, payload)
	if err != nil {
		return
	}
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		log.Println("NewEventEmitter error: ", err)
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	jsonPayload, _ := json.MarshalIndent(payload, "", "\t")

	err = emitter.Push(string(jsonPayload), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logViaRPC(w http.ResponseWriter, entry LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	rpcPayload := RPCPayload{
		Name: entry.Name,
		Data: entry.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: result,
	}

	tools.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) logViagRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := tools.ReadJson(w, r, &requestPayload)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	conn, err := grpc.Dial(
		"logger-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: "logged via gRPC",
	}

	tools.WriteJson(w, http.StatusAccepted, payload)
}
