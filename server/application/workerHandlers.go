package application

import (
	"benefitsDomain/domain/person/personRoles"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (a *Application) addWorkerHandlers() {
	a.router.HandleFunc("/api/workers", a.getWorkers).Methods("GET")
	a.router.HandleFunc("/api/workers/{id}", a.getWorker).Methods("GET")
	a.router.HandleFunc("/api/workers", a.insertWorker).Methods("POST")
	a.router.HandleFunc("/api/workers", a.deleteAllWorkers).Methods("DELETE")

}
func (a *Application) getWorkers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var id, queryType string
	wherePredicate, ok := vars["where"]
	var config map[string]string
	if ok {
		config = make(map[string]string)
		parts := strings.Split(wherePredicate, "=")
		id = parts[1]
		if parts[0] == "type" {
			queryType = "type"
			config["type"] = "type"
			config["id"] = id

		} else if parts[0] == "person" {
			queryType = "person"
			config["type"] = "person"
			config["id"] = id

		} else {
			queryType = "unknown"
		}

	} else {
		queryType = "All"
		config = nil
	}
	d, _ := a.GetPersonDataStore()
	var workers []*personRoles.Worker
	var err error
	switch queryType {
	case "All":
		workers, err = d.GetWorkers(nil)
	default:
		workers, err = d.GetWorkers(config)
	}
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(workers)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
func (a *Application) getWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	d, _ := a.GetPersonDataStore()
	person, err := d.GetWorker(id, "Person")
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(person)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
func (a *Application) insertWorker(w http.ResponseWriter, r *http.Request) {
	// If the Content-Type header is present, check that it has the value
	// application/json. Note that we parse and normalize the header to remove
	// any additional parameters (like charset or boundary information) and normalize
	// it by stripping whitespace and converting to lowercase before we check the
	// value.
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var p personRoles.Worker
	err := dec.Decode(&p)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our Person struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Worker: %+v", p)

	d, _ := a.GetPersonDataStore()
	err = d.InsertWorker(&p)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
}
func (a *Application) deleteAllWorkers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetPersonDataStore()
	err := d.DeleteAllWorkers()
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}

}
