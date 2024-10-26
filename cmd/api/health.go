package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "Ok",
	}
	if err := writeJSON(w, http.StatusOK, data); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
