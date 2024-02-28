package application

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (a *Application) addParticipantHandlers() {
	a.router.HandleFunc("/api/participants", a.getParticipants).Methods("GET")
	a.router.HandleFunc("/api/participantByInternalId/{id}", a.getParticipantByInternalId).Methods("GET")
	a.router.HandleFunc("/api/participants/{pid}/benefits", a.getParticipantByPersonId).Methods("GET")
	a.router.HandleFunc("/api/participants/{pid}/benefits/{bid}", a.getParticipantByPersonIdAndBenefitId).Methods("GET")

}
func (a *Application) getParticipants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetPersonDataStore()
	config := make(map[string]string)
	config["keyType"] = "All"
	config["PersonId"] = ""
	participants, err := d.GetParticipants(config)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	ev := datatypes.EnvironmentVariables{
		TemplateDirectory: "./templates/",
	}
	var sb strings.Builder
	for _, p := range participants {
		s := p.Report(ev)
		sb.WriteString(s)
	}
	_, err = w.Write([]byte(sb.String()))
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
func (a *Application) getParticipantByInternalId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	config := make(map[string]string)
	config["keyType"] = "InternalId"
	config["ParticipantId"] = id
	d, _ := a.GetPersonDataStore()
	participant, err := d.GetParticipant(config)

	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	rc := businessProcess.ResourceContext{}
	rc.SetEnvironmentVariables(a.environmentVariables)
	s := participant.Report(a.environmentVariables)
	_, err = w.Write([]byte(s))
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
}
func (a *Application) getParticipantByPersonId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetPersonDataStore()
	vars := mux.Vars(r)
	pid := vars["pid"]
	config := make(map[string]string)
	config["keyType"] = "PersonId"
	config["PersonId"] = pid
	participants, err := d.GetParticipants(config)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	rc := businessProcess.ResourceContext{}
	rc.SetEnvironmentVariables(a.environmentVariables)
	var sb strings.Builder
	for _, p := range participants {
		s := p.Report(a.environmentVariables)
		sb.WriteString(s)
	}
	_, err = w.Write([]byte(sb.String()))
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
func (a *Application) getParticipantByPersonIdAndBenefitId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d, _ := a.GetPersonDataStore()
	vars := mux.Vars(r)
	pid := vars["pid"]
	bid := vars["bid"]
	config := make(map[string]string)
	config["keyType"] = "PersonId/BenefitId"
	config["PersonId"] = pid
	config["BenefitId"] = bid
	participant, err := d.GetParticipant(config)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	rc := businessProcess.ResourceContext{}
	rc.SetEnvironmentVariables(a.environmentVariables)
	s := participant.Report(a.environmentVariables)
	_, err = w.Write([]byte(s))
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
}
