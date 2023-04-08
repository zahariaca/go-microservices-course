package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zahariaca/toolbox"
)

var tools toolbox.Tools

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := toolbox.JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = tools.WriteJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSumbission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := tools.ReadJson(w, r, &requestPayload)
	if err != nil {
		tools.ErrorJson(w, err)
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
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
