package commands

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"benefitsDomain/domain/person"
	"benefitsDomain/domain/person/personRoles"
	"fmt"
	"server/application"

	"log"
	"log/slog"
	"math/rand"
	"os"
	"strings"
)

// person.Person
func CreateRandomPerson(options map[string]interface{}) *person.Person {
	person := person.Person{}

	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min

	internalId := fmt.Sprintf("%d", id)
	person.InternalId = internalId
	person.LastName = "Sample_" + internalId
	person.FirstName = "John"
	//person.BirthDate = YYYYMMDD_Date("20230101")
	person.BirthDate = datatypes.YYYYMMDD_Date("20230101")
	o := make(map[string]string)
	person.ContactPreferenceHistory = buildContact(o)

	return &person
}
func buildContact(options map[string]string) []person.ContactPreferenceHistory {
	history := make([]person.ContactPreferenceHistory, 0)
	contactType := [3]string{"Home Address", "Home Email", "Work Email"}
	contactClass := [3]string{"Street", "Email", "Email"}
	for i := 0; i < len(contactType); i++ {
		o := make(map[string]string)
		o["Type"] = contactType[i]
		o["Class"] = contactClass[i]
		var contactPoint person.ContactPoint
		switch contactClass[i] {
		case "Street":
			contactPoint = buildRandomStreetAddress(o)
		case "Email":
			contactPoint = buildRandomEmailAddress(o)

		case "Phone":
			contactPoint = buildRandomPhone(o)

		}
		cpID := contactPoint.GetContactId()
		preferenceItem := person.ContactPreferencePeriod{
			EffectiveBeginDate: "01/01/2023",
			EffectiveEndDate:   "01/01/2024",
			ContactPointId:     cpID,
			ContactPoint:       contactPoint,
		}
		preferenceItems := make([]person.ContactPreferencePeriod, 0)
		preferenceItems = append(preferenceItems, preferenceItem)
		preferenceHistory := person.ContactPreferenceHistory{
			ContactPointType:          contactType[i],
			ContactPointClass:         contactClass[i],
			ContractPreferencePeriods: preferenceItems,
		}
		history = append(history, preferenceHistory)
	}
	return history
}
func buildRandomStreetAddress(options map[string]string) person.StreetAddress {
	s := person.StreetAddress{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.AddressLine1 = "100 Main St"
	s.City = "Chicago"
	s.State = "IL"
	s.Zipcode = "12345"
	return s
}
func buildRandomPhone(options map[string]string) person.PhoneNumber {
	s := person.PhoneNumber{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.PhoneNumber = "(123) 456-7890)"

	return s
}
func buildRandomEmailAddress(options map[string]string) person.EmailAddress {
	s := person.EmailAddress{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.EmailAddress = "sample@company.com"
	return s
}
func WritePersonToDisk(p *person.Person, iFileName string) {
	fileName := iFileName
	d1, err := person.CreateJsonFromPerson(p)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fileName, d1, 0644)
	if err != nil {
		panic(err)
	}
}
func LoadPersonsFromDirectory() {
	slog.Debug("Loading people from directory")
	directory := "../MockData/"
	files, err := os.ReadDir(directory)

	if err != nil {
		slog.Error("Error", err)
	}

	personMongoDB := db.NewPersonMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	defer personMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	// CORS is enabled only in prod profile
	ev := datatypes.EnvironmentVariables{
		TemplateDirectory: "./templates/",
	}
	ev.Cors = os.Getenv("profile") == "prod"
	app := application.NewApplication(personMongoDB, businessProcessMongoDB, ev)
	d, _ := app.GetPersonDataStore()
	d.DeleteAllPersons()

	for _, f := range files {
		if strings.Contains(f.Name(), "Person") {
			name := directory + f.Name()
			person, _ := person.CreatePersonFromJsonFile(name)
			err = d.InsertPerson(person)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
func CreateRandomPersonToDisk(num int) {
	options := make(map[string]interface{})

	for i := 0; i < num; i++ {

		person := CreateRandomPerson(options)
		filename := "../MockData/Person_" + person.InternalId + ".json"
		WritePersonToDisk(person, filename)
	}
}

// Participant
func CreateRandomParticipant(options map[string]interface{}, person *person.Person) *personRoles.Participant {
	participant := personRoles.Participant{}

	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min

	internalId := fmt.Sprintf("%d", id)
	participant.InternalId = internalId
	participant.PersonId = person.InternalId
	participant.BenefitId = "Medical"

	return &participant
}

func WriteParticipantToDisk(participant *personRoles.Participant, iFileName string) {
	fileName := iFileName
	d1, err := personRoles.CreateJsonFromParticipant(participant)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fileName, d1, 0644)
	if err != nil {
		panic(err)
	}
}
func LoadParticipantsFromDirectory(createParticipants bool) {
	slog.Debug("Loading participants from directory")
	directory := "../MockData/"
	files, err := os.ReadDir(directory)

	if err != nil {
		slog.Error("Error", err)
	}

	personMongoDB := db.NewPersonMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	defer personMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	// CORS is enabled only in prod profile
	ev := datatypes.EnvironmentVariables{
		TemplateDirectory: "./templates/",
	}
	ev.Cors = os.Getenv("profile") == "prod"
	app := application.NewApplication(personMongoDB, businessProcessMongoDB, ev)
	d, _ := app.GetPersonDataStore()
	d.DeleteAllParticipants()
	if !createParticipants {
		return
	}
	for _, f := range files {
		name := directory + f.Name()
		if strings.Contains(f.Name(), "Participant") {
			participant, _ := personRoles.CreateParticipantFromJsonFile(name)
			err = d.InsertParticipant(participant)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
func CreateRandomParticipantToDisk() {
	options := make(map[string]interface{})
	directory := "../MockData/"
	files, err := os.ReadDir(directory)
	if err != nil {
		slog.Error("Error", err)
	}
	for _, f := range files {
		name := directory + f.Name()
		if strings.Contains(f.Name(), "Person") {
			prsn, _ := person.CreatePersonFromJsonFile(name)
			p := CreateRandomParticipant(options, prsn)
			filename := "../MockData/Participant" + p.InternalId + ".json"
			WriteParticipantToDisk(p, filename)
		}
	}
}

//Worker

func CreateRandomWorker(options map[string]interface{}, person *person.Person) *personRoles.Worker {
	worker := personRoles.Worker{}

	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min

	internalId := fmt.Sprintf("%d", id)
	worker.InternalId = internalId
	worker.WorkerId = internalId
	worker.PersonInternalId = person.InternalId
	worker.Employer = "Acme"

	return &worker
}

func WriteWorkerToDisk(w *personRoles.Worker, iFileName string) {
	fileName := iFileName
	d1, err := personRoles.CreateJsonFromWorker(w)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fileName, d1, 0644)
	if err != nil {
		panic(err)
	}
}
func LoadWorkersFromDirectory() {
	slog.Debug("Loading workers from directory")
	directory := "../MockData/"
	files, err := os.ReadDir(directory)

	if err != nil {
		slog.Error("Error", err)
	}

	personMongoDB := db.NewPersonMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	defer personMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	// CORS is enabled only in prod profile
	ev := datatypes.EnvironmentVariables{
		TemplateDirectory: "./templates/",
	}
	ev.Cors = os.Getenv("profile") == "prod"
	app := application.NewApplication(personMongoDB, businessProcessMongoDB, ev)
	d, _ := app.GetPersonDataStore()
	d.DeleteAllWorkers()

	for _, f := range files {
		name := directory + f.Name()
		if strings.Contains(f.Name(), "Worker") {
			w, _ := personRoles.CreateWorkerFromJsonFile(name)
			err = d.InsertWorker(w)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
func CreateRandomWorkersToDisk(num int) {
	options := make(map[string]interface{})
	directory := "../MockData/"
	files, err := os.ReadDir(directory)
	if err != nil {
		slog.Error("Error", err)
	}
	for _, f := range files {
		name := directory + f.Name()
		if strings.Contains(f.Name(), "domain.Person") {
			prsn, _ := person.CreatePersonFromJsonFile(name)
			p := CreateRandomWorker(options, prsn)
			filename := "../MockData/Worker" + p.InternalId + ".json"
			WriteWorkerToDisk(p, filename)
		}
	}
}

/*
func createBusinessProcessDefinitionGraph() {
	businessProcess := domain.CreateMockBusinessProcessDefinitionObjects("BP001")
	flow := businessProcess.Flow
	s := graphs.BuildBusinessProcessDefinitionDotString(&flow)
	fileName := "../graphs/openEnrollment.dot"

	err := os.WriteFile(fileName, []byte(s), 0644)
	if err != nil {
		panic(err)
	}

}

func createBusinessProcessDefinitionReport() {
	businessProcess := domain.CreateMockBusinessProcessDefinitionObjects("BP001")
	s := businessProcess.Report()
	fmt.Println(s)
}
*/
