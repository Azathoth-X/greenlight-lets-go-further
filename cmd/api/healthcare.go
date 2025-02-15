package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"env":     app.config.env,
			"version": version,
		},
	}

	err := app.writeJson(w, http.StatusOK, data, nil)

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "somehting went wrong while handling request", http.StatusInternalServerError)
	}
}
