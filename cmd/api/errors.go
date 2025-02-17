package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "url", uri)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {

	msg := envelope{
		"message": message,
	}

	err := app.writeJson(w, status, msg, nil)

	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "The server couldnt handle your request"
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {

	msg := "The resource requested can not be found"
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {

	msg := fmt.Sprintf("%s request on this route is not allowed", r.Method)
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}
