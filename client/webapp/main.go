// forms.go
package main

import (
	"bytes"
	"client/webapp/forms"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
var SignalCh = make(chan os.Signal, 1)

func main() {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	/*
		ev := datatypes.EnvironmentVariables{
			TemplateDirectory: "./templates/",
		}
	*/
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	// Create a context to handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Create a WaitGroup to keep track of running goroutines
	var wg sync.WaitGroup

	// Start the HTTP server
	wg.Add(1)
	go startHTTPServer(ctx, &wg)

	// Listen for termination signals

	signal.Notify(SignalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-SignalCh

	// Start the graceful shutdown process
	slog.Info("Gracefully shutting down HTTP server...")

	// Cancel the context to signal the HTTP server to stop
	cancel()

	// Wait for the HTTP server to finish
	wg.Wait()

	slog.Info("Shutdown complete.")
}

func startHTTPServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	Templ1 = template.Must(template.ParseFiles("./public/electionForm.html"))
	Templ2 = template.Must(template.ParseFiles("./public/eventForm.html"))
	Templ3 = template.Must(template.ParseFiles("./public/jobSubmissionForm.html"))

	forms.InitializePerson()
	fileServer := http.FileServer(http.Dir("./public"))
	http.Handle("/", fileServer)
	http.HandleFunc("/election", handleElectionForm)
	http.HandleFunc("/event", handleEventForm)
	http.HandleFunc("/job", handleJobSubmissionForm)
	http.HandleFunc("/admin/shutdown", handleShutdown)

	server := &http.Server{
		Addr:    ":3000",
		Handler: nil,
	}

	// Start the HTTP server in a separate goroutine
	go func() {
		slog.Info("Starting HTTP server on Port 3000..")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %s\n", err)
		}
	}()

	// Wait for the context to be canceled
	select {
	case <-ctx.Done():
		// Shutdown the server gracefully
		slog.Info("<- ctx.Done in startHTTPServer-- Shutting down HTTP server gracefully...")
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			slog.Info("HTTP server shutdown error: %s\n", err)
		}
	}

	slog.Info("HTTP server stopped.")
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

	slog.Info("response Status:" + response.Status)
	s := fmt.Sprintf("response Header: %v", response.Header)
	slog.Info(s)
	body, _ := io.ReadAll(response.Body)
	slog.Info("response Body:" + string(body))

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

	slog.Info("response Status:" + response.Status)
	s := fmt.Sprintf("response Headers: %v", response.Header)
	slog.Info(s)
	body, _ := io.ReadAll(response.Body)
	slog.Info("response Body:" + string(body))

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

	slog.Info("response Status:" + response.Status)
	s := fmt.Sprintf("response Headers: %v", response.Header)
	slog.Info(s)
	body, _ := io.ReadAll(response.Body)
	slog.Info("response Body:" + string(body))

	_, err3 := w.Write(body)
	if err3 != nil {
		log.Fatal(err3)
	}
}
func determinePersonBusinessReferenceNumber(key string) string {
	id := fmt.Sprintf("0%s-%s-00%s_BP001_20240101", key, key, key)
	return id
}
func handleShutdown(w http.ResponseWriter, r *http.Request) {
	SignalCh <- syscall.SIGTERM
}
