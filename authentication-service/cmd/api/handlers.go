package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/zahariaca/toolbox"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var tools toolbox.Tools

	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := tools.ReadJson(w, r, &requestPayload)

	if err != nil {
		tools.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	// validate the user agains the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		tools.ErrorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		tools.ErrorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	tools.WriteJson(w, http.StatusAccepted, payload)
}
