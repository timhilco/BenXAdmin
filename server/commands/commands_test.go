package commands

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/db"
	"benefitsDomain/domain/person"
	"fmt"

	"log/slog"
	"os"
	"testing"
)

func TestLoad_Benefit(t *testing.T) {
	fileName := "./client/A1234/BenefitPlan_Rates_01012024.xlsx"
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	slog.Debug("Starting Benefit Spreadsheet Loader")
	planMongoDB := db.NewPlanMongo()
	defer planMongoDB.CloseClientDB()
	err := planMongoDB.DeleteAllBenefits()
	if err != nil {
		fmt.Println(" Error in delete: " + err.Error())
	}
	_, benefits, _ := LoadBenefitFromSpreadsheet(fileName)
	for _, benefit := range benefits {
		planMongoDB.InsertBenefit(benefit)
	}
}
func TestPerson_CompleteDatabaseRefresh(t *testing.T) {
	createParticipants := false
	directory := "../MockData/"
	files, err := os.ReadDir(directory)
	if err != nil {
		slog.Error("Error", err)
	}

	for _, f := range files {
		name := directory + f.Name()
		err = os.Remove(name)
		if err != nil {
			slog.Error("Error", err)
		}
	}
	num := 10
	CreateRandomPersonToDisk(num)
	LoadPersonsFromDirectory()
	if createParticipants {
		CreateRandomParticipantToDisk()
	}
	LoadParticipantsFromDirectory(createParticipants)
	num = 10
	CreateRandomWorkersToDisk(num)
	LoadWorkersFromDirectory()
}
func TestPerson_CreateRandomPersonToDisk(t *testing.T) {
	num := 10
	CreateRandomPersonToDisk(num)
}
func TestPerson_MongoLoadPersonsFromDirectory(t *testing.T) {
	LoadPersonsFromDirectory()
}
func TestPerson_CreateRandomParticipantToDisk(t *testing.T) {
	CreateRandomParticipantToDisk()
}
func TestPerson_MongoLoadParticipantsFromDirectory(t *testing.T) {
	LoadParticipantsFromDirectory(true)
}
func TestPerson_CreateRandomWorkerToDisk(t *testing.T) {
	num := 10
	CreateRandomWorkersToDisk(num)
}
func TestPerson_MongoLoadWorkerFromDirectory(t *testing.T) {
	LoadWorkersFromDirectory()
}
func Test_BenefitProcessDefinitionGraph(t *testing.T) {
	//createBusinessProcessDefinitionGraph()
}
func Test_BenefitProcessDefinitionReport(t *testing.T) {
	//createBusinessProcessDefinitionReport()
}
func Test_Person_Report(t *testing.T) {
	person := person.Person{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	ev := datatypes.EnvironmentVariables{}
	r := person.Report(ev)
	fmt.Println(r)
}
