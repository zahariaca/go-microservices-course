package main

import (
	"fmt"
	"github.com/zahariaca/toolbox"
	"log"
	"net/http"
)

var tools toolbox.Tools

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering SendEmail...")
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	err := tools.ReadJson(w, r, &requestPayload)
	if err != nil {
		tools.ErrorJson(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	log.Println("msg before SendSMTPMessage:", msg)
	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		log.Println("After SendSMTPMessage error:", err)
		tools.ErrorJson(w, err)
		return
	}

	payload := toolbox.JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("sent to:%s", requestPayload.To),
	}

	tools.WriteJson(w, http.StatusAccepted, payload)
}
