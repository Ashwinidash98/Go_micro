package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type RequestPayLoad struct {
	Action string      `json:"action"`
	Auth   AuthPayLoad `json:"auth,omitempty"`
}

type AuthPayLoad struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayLoad RequestPayLoad

	err := app.readJSON(w, r, &requestPayLoad)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayLoad.Action {
	case "auth":
		app.authenticate(w, requestPayLoad.Auth)

	default:
		app.errorJSON(w, errors.New("Unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayLoad) {
	//create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	//call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// client := &http.Client{}
	// response, err := client.Do(request)
	// if err != nil {
	// 	app.errorJSON(w, err)
	// 	return
	// }

	// defer response.Body.Close()
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	log.Println("Status:", response.Status)

	body, _ := io.ReadAll(response.Body)
	log.Println("Body:", string(body))

	// Reset the body so it can be decoded later if needed
	response.Body = io.NopCloser(bytes.NewBuffer(body))

	//make sure we get back the correct statsu code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	//create a variable we'll read Response.body into
	var jsonFromService jsonResponse

	//decode the json from the auth servoce
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}
