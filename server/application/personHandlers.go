package application

import (
	"benefitsDomain/apiResponse"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/person"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (a *Application) addPersonHandlers() {
	a.router.HandleFunc("/api/persons", a.getPersons).Methods("GET")
	a.router.HandleFunc("/api/persons/{id}", a.getPerson).Methods("GET")
	a.router.HandleFunc("/api/persons/{id}/view/{viewId}", a.getPersonView).Methods("GET")
	a.router.HandleFunc("/api/persons", a.insertPerson).Methods("POST")
	a.router.HandleFunc("/api/persons", a.deleteAllPersons).Methods("DELETE")

}
func (a *Application) getPersons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetPersonDataStore()
	persons, err := d.GetPersons()
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(persons)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
func (a *Application) getPerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	d, _ := a.GetPersonDataStore()
	person, err := d.GetPerson(id, "External")
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	ev := (a.environmentVariables)
	s := person.Report(ev)
	_, err = w.Write([]byte(s))
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
func (a *Application) getPersonView(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	viewId := vars["viewId"]
	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetPersonDataStore()
	person, err := d.GetPerson(id, "External")
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	switch viewId {
	case "profile":
		personView := apiResponse.PersonProfileViewResponse{
			InternalId: person.InternalId,
			ExternalId: person.ExternalId,
			FirstName:  person.FirstName,
			LastName:   person.LastName,
			BirthDate:  person.BirthDate.FormattedString(""),
		}
		config := make(map[string]string)
		config["keyType"] = "PersonId"
		config["PersonId"] = person.InternalId
		workers, err := d.GetWorkers(config)
		if err != nil {
			if err.Error() != "worker not found" {
				sendErr(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		ws := make([]apiResponse.Worker, 0)
		for _, w := range workers {
			w := apiResponse.Worker{
				InternalId:       w.InternalId,
				WorkerId:         w.WorkerId,
				Employer:         w.Employer,
				Pay:              w.Pay.FormattedString(""),
				EmploymentStatus: w.EmploymentStatus,
			}
			ws = append(ws, w)
		}
		personView.Workers = ws
		participants, err := d.GetParticipants(config)
		if err != nil {
			if err.Error() != "participant not found" {
				sendErr(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		ps := make([]apiResponse.Participant, 0)
		for _, p := range participants {
			ap := apiResponse.Participant{
				BenefitId:  p.BenefitId,
				InternalId: p.InternalId,
				PersonId:   p.PersonId,
			}
			ch := p.CoverageHistory.CoveragePeriods[0]
			ap.CoverageStartDate = ch.CoverageStartDate.FormattedString("")
			ap.CoverageEndDate = ch.CoverageEndDate.FormattedString("")
			ap.ElectedCoverageLevel = ch.ElectedCoverageLevel
			ap.PayrollReportingState = ch.PayrollReportingState
			ap.CarrierReportingState = ch.CarrierReportingState
			ap.ElectedBenefitOfferingId = ch.ElectedBenefitOfferingId
			ap.ActualBenefitOfferingId = ch.ActualBenefitOfferingId
			ap.ElectedCoverageLevel = ch.ElectedCoverageLevel
			ap.ActualCoverageLevel = ch.ActualCoverageLevel
			ap.OfferedBenefitOfferingId = ch.OfferedBenefitOfferingId
			ap.ActualCoverageAmount = ch.ActualCoverageAmount
			ap.ElectedCoverageAmount = ch.ElectedCoverageAmount
			ap.EmployeePreTaxCost = ch.EmployeePreTaxCost
			ap.EmployerCost = ch.EmployerCost
			ap.EmployeeAfterTaxCost = ch.EmployeeAfterTaxCost
			ap.EmployerSubsidy = ch.EmployerSubsidy
			ps = append(ps, ap)
		}
		personView.Participants = ps
		d2, _ := a.GetBusinessProcessDataStore()
		config = make(map[string]string)
		config["type"] = "person"
		config["id"] = person.InternalId

		businessProcesses, _ := d2.GetPersonBusinessProcesses(config)
		bps := make([]apiResponse.BusinessProcess, 0)
		for _, bp := range businessProcesses {
			b := apiResponse.BusinessProcess{
				InternalId:                  bp.InternalId,
				ReferenceNumber:             bp.ReferenceNumber,
				BusinessProcessDefinitionId: bp.BusinessProcessDefinitionId,
				EffectiveDate:               bp.EffectiveDate.FormattedString(""),
			}
			switch bp.State {
			case businessProcess.C_STATE_OPEN:
				b.State = "Open"
			case businessProcess.C_STATE_CLOSED:
				b.State = "Closed"
			}
			bps = append(bps, b)
		}
		personView.BusinessProcesses = bps
		err = json.NewEncoder(w).Encode(personView)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	}
}
func (a *Application) insertPerson(w http.ResponseWriter, r *http.Request) {
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

	var p person.Person
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

	fmt.Fprintf(w, "Person: %+v", p)

	d, _ := a.GetPersonDataStore()
	err = d.InsertPerson(&p)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
}
func (a *Application) deleteAllPersons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetPersonDataStore()
	err := d.DeleteAllPersons()
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}

}
