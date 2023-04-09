package main

import (
	"github.com/zahariaca/logger-service/data"
	"github.com/zahariaca/toolbox"
	"net/http"
)

var tools toolbox.Tools

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	// read json into var
	var requestPayload JSONPayload
	_ = tools.ReadJson(w, r, &requestPayload)

	// insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	resp := toolbox.JsonResponse{
		Error:   false,
		Message: "logged",
	}

	tools.WriteJson(w, http.StatusAccepted, resp)
}
