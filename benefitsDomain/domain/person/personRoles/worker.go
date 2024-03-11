package personRoles

import (
	"benefitsDomain/datatypes"
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"os"
)

type Worker struct {
	InternalId           string             `json:"internalId" bson:"internalId"`
	WorkerId             string             `json:"workerId" bson:"workerId"`
	PersonInternalId     string             `json:"personInternalId" bson:"personInternalId"`
	Employer             string             `json:"employer" bson:"employer"`
	Pay                  datatypes.BigFloat `json:"pay" bson:"pay"`
	EmploymentStatus     string             `json:"employmentStatus" bson:"employmentStatus"`
	EmploymentCategories map[string]string  `json:"employmentCategories" bson:"employmentCategories"`
	Organizations        map[string]string  `json:"organizations" bson:"organizations"`
}

func CreateJsonFromWorker(worker *Worker) ([]byte, error) {
	// Convert struct to JSON
	jsonData, err := json.Marshal(worker)
	if err != nil {
		log.Fatal(err)
	}
	return jsonData, err
}
func CreateWorkerFromJsonFile(filename string) (*Worker, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	worker := Worker{}

	err = json.Unmarshal(data, &worker) // here!
	if err != nil {
		panic(err)
	}
	return &worker, nil
}
func (w *Worker) GetCurrentPay() (datatypes.BigFloat, error) {
	return w.Pay, nil
}
func (w *Worker) Report(ev datatypes.EnvironmentVariables) string {
	dir := ev.TemplateDirectory
	templateFile := dir + "workerTemplate.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, w)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
