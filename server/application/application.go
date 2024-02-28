package application

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"encoding/json"

	"log/slog"
	"net/http"
	batchjobs "server/batchJobs"

	"github.com/gorilla/mux"
)

type Application struct {
	personDataStore          *db.PersonMongoDB
	businessProcessDataStore *businessProcess.BusinessProcessMongoDB
	router                   *mux.Router
	environmentVariables     datatypes.EnvironmentVariables
	resourceContext          *businessProcess.ResourceContext
}

func (a *Application) Serve() error {
	slog.Info("Web server is available on port 8080")
	return http.ListenAndServe(":8080", a.router)
}
func (a *Application) GetPersonDataStore() (*db.PersonMongoDB, error) {
	return a.personDataStore, nil
}
func (a *Application) GetBusinessProcessDataStore() (*businessProcess.BusinessProcessMongoDB, error) {
	return a.businessProcessDataStore, nil
}

func sendErr(w http.ResponseWriter, code int, message string) {
	resp, _ := json.Marshal(map[string]string{"error": message})
	http.Error(w, string(resp), code)
}
func NewApplication(p *db.PersonMongoDB, bp *businessProcess.BusinessProcessMongoDB, ev datatypes.EnvironmentVariables) Application {
	r := mux.NewRouter()
	app := Application{
		personDataStore:          p,
		businessProcessDataStore: bp,
		router:                   r,
		environmentVariables:     ev,
	}
	app.addPersonHandlers()
	app.addParticipantHandlers()
	app.addWorkerHandlers()
	app.addBenefitProcessHandlers()
	app.addJobSubmissionHandlers()
	return app
}
func NewApplication2(personMongoDB *db.PersonMongoDB, businessProcessMongoDB *businessProcess.BusinessProcessMongoDB, planMongoDB *db.PlanMongoDB, ev datatypes.EnvironmentVariables) Application {
	r := mux.NewRouter()

	numberOfWorkers := 1
	var ed businessProcess.MessageBroker
	switch ev.IsKafka {
	case true:
		configMap := make(map[string]string)
		configMap["client.id"] = "BenX"
		configMap["group.id"] = "Hilco1"
		ed = businessProcess.NewKafkaProducerMessageBroker(personMongoDB, businessProcessMongoDB, configMap)
	default:
		ed = businessProcess.NewChannelMessageBroker(personMongoDB, businessProcessMongoDB, numberOfWorkers)
	}
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	rc := businessProcess.NewResourceContext("Application::updatePersonBusinessProcessEvent", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)

	app := Application{
		personDataStore:          personMongoDB,
		businessProcessDataStore: businessProcessMongoDB,
		router:                   r,
		environmentVariables:     ev,
		resourceContext:          rc,
	}
	app.addPersonHandlers()
	app.addParticipantHandlers()
	app.addWorkerHandlers()
	app.addBenefitProcessHandlers()
	app.addJobSubmissionHandlers()
	return app
}

// Needed in order to disable CORS for local development
/*
func disableCors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		h(w, r)
	}
}
*/
func StartConsumer(
	personMongoDB *db.PersonMongoDB, businessProcessMongoDB *businessProcess.BusinessProcessMongoDB, planMongoDB *db.PlanMongoDB) {
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ev := datatypes.EnvironmentVariables{}
	configMap := make(map[string]string)
	configMap["client.id"] = "BenX"
	configMap["group.id"] = "Hilco1"
	batchjobs.StartKafkaMessageConsumer(personMongoDB, businessProcessMongoDB, configMap, bpd, planMongoDB, ev)
}
