// forms.go
package main

import (
	"bytes"
	"client/webapp/forms"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
)

type OpenEnrollmentElectionRequest struct {
	BenefitPlanElections []OpenEnrollmentElection `json:"benefitPlanElections"`
}
type OpenEnrollmentElection struct {
	BenefitId       string `json:"benefitId"`
	BenefitPlanId   string `json:"benefitPlanId"`
	CoverageLevelId string `json:"coverageLevelId"`
	CoverageAmount  string `json:"coverageAmount"`
}
type Event struct {
	EventId string `json:"eventId"`
	IsKafka string `json:"isKafka"`
}
type Job struct {
	JobName string `json:"jobName"`
}

var Templ1 *template.Template
var Templ2 *template.Template
var Templ3 *template.Template

func main() {
	Templ1 = template.Must(template.ParseFiles("./public/electionForm.html"))
	Templ2 = template.Must(template.ParseFiles("./public/eventForm.html"))
	Templ3 = template.Must(template.ParseFiles("./public/jobSubmissionForm.html"))

	forms.InitializePerson()
	fileServer := http.FileServer(http.Dir("./public"))
	http.Handle("/", fileServer)
	http.HandleFunc("/election", handleElectionForm)
	http.HandleFunc("/event", handleEventForm)
	http.HandleFunc("/job", handleJobSubmissionForm)

	slog.Info("Web server is available on port 3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func handleElectionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err0 := Templ1.Execute(w, nil)
		if err0 != nil {
			log.Fatal(err0)
		}
		return
	}

	medical := OpenEnrollmentElection{
		BenefitId:       r.FormValue("mBenefitId"),
		BenefitPlanId:   r.FormValue("mBenefitPlanId"),
		CoverageLevelId: r.FormValue("mCoverageLevelId"),
		CoverageAmount:  "0",
	}
	dental := OpenEnrollmentElection{
		BenefitId:       r.FormValue("dBenefitId"),
		BenefitPlanId:   r.FormValue("dBenefitPlanId"),
		CoverageLevelId: r.FormValue("dCoverageLevelId"),
		CoverageAmount:  "0",
	}
	life := OpenEnrollmentElection{
		BenefitId:       r.FormValue("elBenefitId"),
		BenefitPlanId:   r.FormValue("elBenefitPlanId"),
		CoverageLevelId: r.FormValue("elCoverageLevelId"),
		CoverageAmount:  "0",
	}
	hsa := OpenEnrollmentElection{
		BenefitId:       r.FormValue("hcBenefitId"),
		BenefitPlanId:   r.FormValue("hcBenefitPlanId"),
		CoverageLevelId: "0",
		CoverageAmount:  r.FormValue("hcContribution"),
	}
	dcsa := OpenEnrollmentElection{
		BenefitId:       r.FormValue("dcBenefitId"),
		BenefitPlanId:   r.FormValue("dcBenefitPlanId"),
		CoverageLevelId: "0",
		CoverageAmount:  r.FormValue("dcContribution"),
	}
	be := make([]OpenEnrollmentElection, 0)
	be = append(be, medical, dental, life, hsa, dcsa)

	elections := OpenEnrollmentElectionRequest{
		BenefitPlanElections: be,
	}

	jsonData, err2 := json.Marshal(elections)
	if err2 != nil {
		log.Fatal(err2)
	}
	sid := r.FormValue("personId")
	id := determinePersonBusinessReferenceNumber(sid)
	url := "http://localhost:8080/api/personBusinessProcesses/" + id + "/elections"
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		log.Fatal(error)
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))

	_, err3 := w.Write(body)
	if err3 != nil {
		log.Fatal(err3)
	}
}
func handleEventForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err0 := Templ2.Execute(w, nil)
		if err0 != nil {
			log.Fatal(err0)
		}
		return
	}
	eventId := r.FormValue("eventId")
	isKafka := r.FormValue("isKafka")
	event := Event{
		EventId: eventId,
	}
	if isKafka == "checked" {
		event.IsKafka = "true"
	} else {
		event.IsKafka = "false"

	}
	jsonData, err2 := json.Marshal(event)
	if err2 != nil {
		log.Fatal(err2)
	}
	sid := r.FormValue("referenceNumber")

	id := determinePersonBusinessReferenceNumber(sid)
	url := "http://localhost:8080/api/personBusinessProcesses/" + id + "/event"
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		log.Fatal(error)
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))

	_, err3 := w.Write(body)
	if err3 != nil {
		log.Fatal(err3)
	}
}
func handleJobSubmissionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err0 := Templ3.Execute(w, nil)
		if err0 != nil {
			log.Fatal(err0)
		}
		return
	}
	isAnnualEnrollmentJob := r.FormValue("isAnnualEnrollmentJob")
	var jobName string
	if isAnnualEnrollmentJob == "checked" {
		jobName = "AnnualEnrollmentStart"
	}
	job := Job{
		JobName: jobName,
	}
	jsonData, err2 := json.Marshal(job)
	if err2 != nil {
		log.Fatal(err2)
	}

	url := "http://localhost:8080/api/jobSubmission"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		log.Fatal(error)
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))

	_, err3 := w.Write(body)
	if err3 != nil {
		log.Fatal(err3)
	}
}
func determinePersonBusinessReferenceNumber(key string) string {
	id := fmt.Sprintf("0%s-%s-00%s_BP001_20240101", key, key, key)
	return id
}
