package application

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"server/message"

	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/spaolacci/murmur3"
)

type Event struct {
	EventId string `json:"eventId"`
	IsKafka string `json:"isKafka"`
}

func (a *Application) addBenefitProcessHandlers() {
	a.router.HandleFunc("/api/personBusinessProcesses/{id}", a.getPersonBusinessProcessReport).Methods("GET")
	a.router.HandleFunc("/api/personBusinessProcesses", a.getPersonBusinessProcessCollection).Methods("GET")
	a.router.HandleFunc("/api/personBusinessProcesses/{type}", a.getPersonBusinessProcessesByType).Methods("GET")
	a.router.HandleFunc("/api/personBusinessProcesses/{id}/elections", a.updatePersonBusinessProcessElections).Methods("PUT")
	a.router.HandleFunc("/api/personBusinessProcesses/{id}/event", a.updatePersonBusinessProcessEvent).Methods("PUT")

}

func (a *Application) getPersonBusinessProcessReport(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	d, _ := a.GetBusinessProcessDataStore()
	pbp, err := d.GetPersonBusinessProcess(id)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	rc := businessProcess.ResourceContext{}
	rc.SetEnvironmentVariables(a.environmentVariables)
	s := pbp.Report(&rc)
	_, err = w.Write([]byte(s))
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}

}
func (a *Application) getPersonBusinessProcessCollection(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetBusinessProcessDataStore()
	pbp, err := d.GetPersonBusinessProcesses()
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(pbp)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}

}
func (a *Application) getPersonBusinessProcessesByType(w http.ResponseWriter, r *http.Request) {
	/*
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id := vars["id"]
		businessProcess := db.CreateMockBusinessProcessDefinitionObjects(id)
		s := businessProcess.Report()
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(s))
	*/

}
func (a *Application) updatePersonBusinessProcessElections(w http.ResponseWriter, r *http.Request) {
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

	var p businessProcess.OpenEnrollmentElectionRequest
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

	//pbp.ProcessMessage(rc *ResourceContext, event message.Message) {
	vars := mux.Vars(r)
	id := vars["id"]
	d, _ := a.GetBusinessProcessDataStore()
	pbp, err := d.GetPersonBusinessProcess(id)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	personMongoDB := db.NewPersonMongo()
	planMongoDB := db.NewPlanMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	defer personMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	defer planMongoDB.CloseClientDB()

	numberOfWorkers := 1
	ed := businessProcess.NewChannelMessageBroker(personMongoDB, businessProcessMongoDB, numberOfWorkers)
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ev := datatypes.EnvironmentVariables{
		TemplateDirectory: "./templates/",
	}
	rc := businessProcess.NewResourceContext("Application::updatePersonBusinessProcessElections", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)
	anEvent := BuildCommand("PC0001", id, p)
	pbp.ProcessMessage(rc, anEvent)
	res := pbp.Report(rc)
	fmt.Fprintln(w, res)

}
func (a *Application) updatePersonBusinessProcessEvent(w http.ResponseWriter, r *http.Request) {
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

	var evt Event
	err := dec.Decode(&evt)
	eventId := evt.EventId
	//isKafka := evt.IsKafka
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

	//pbp.ProcessMessage(rc *ResourceContext, event message.Message) {
	vars := mux.Vars(r)
	id := vars["id"]

	anEvent := message.BuildEvent(eventId, id)
	ctx := context.Background()
	rc := a.resourceContext
	ed := rc.GetMessageBroker()
	hasher := murmur3.New128()
	hasher.Write([]byte(id))
	partition, _ := hasher.Sum128()
	partition = partition % uint64(10)
	ed.Publish(ctx, anEvent, int(partition))
	slog.Info("Sleeping 5 seconds ...")
	time.Sleep(5 * time.Second)
	d, _ := a.GetBusinessProcessDataStore()
	pbp, err := d.GetPersonBusinessProcess(id)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	res := pbp.Report(rc)
	fmt.Fprintln(w, res)
}
func BuildCommand(id string, referenceNumber string, elections businessProcess.OpenEnrollmentElectionRequest) message.Message {
	dfn := message.CreateMockCommandDefinitionObject(id)
	header := message.CommandHeader{
		CommandId:   id,
		CommandName: dfn.Name,
	}
	domain := message.CommandData{
		ReferenceNumber: referenceNumber,
		Target:          "This",
	}

	jsonData, _ := json.Marshal(elections)
	domain.JsonData = string(jsonData)

	aCommand := message.Command{
		Header: header,
		Data:   domain,
	}
	return &aCommand

}
