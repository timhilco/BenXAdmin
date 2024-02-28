package scenarios

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"benefitsDomain/domain/person"
	"context"
	"encoding/json"
	"fmt"
	"server/kafka"
	"server/message"

	"sync"

	"os"
	"os/signal"
	"syscall"
	"testing"

	"log/slog"

	"github.com/spaolacci/murmur3"
)

func Test_EventDistributor_Scenario1(t *testing.T) {
	refNum := "UNKNOWN"
	events := buildEnrollmentEventSequence1(refNum)
	anEvent := message.BuildCommand("PC0002", "001_BP002_20240101")
	events = append(events, anEvent)
	processEvents(events)
}
func Test_EventDistributor_Scenario2(t *testing.T) {
	refNum := "UNKNOWN"
	events := buildEnrollmentEventSequence2(refNum)
	processEvents(events)
}
func processEvents(events []message.Message) {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	slog.Debug("Starting Event Distributor")
	personMongoDB := db.NewPersonMongo()
	planMongoDB := db.NewPlanMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	defer personMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	defer planMongoDB.CloseClientDB()

	ctx := context.TODO()
	numberOfWorkers := 1
	ed := businessProcess.NewChannelMessageBroker(personMongoDB, businessProcessMongoDB, numberOfWorkers)
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ev := datatypes.EnvironmentVariables{}
	rc := businessProcess.NewResourceContext("scenarios::processEvents", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)
	population, _ := personMongoDB.GetPersons()
	effectiveDate := datatypes.YYYYMMDD_Date("20240101")
	businessProcessDefinition := bpd.GetBusinessProcessDefinition("BP001")
	//businessProcessDefinition.Report()
	businessProcessMongoDB.DeleteAllPersonBusinessProcesses()
	personMongoDB.DeleteAllParticipants()
	var refNum string
	slog.Debug("********************************")
	slog.Debug("Starting Enrollment Event")
	slog.Debug("********************************")
	person := getFirstPerson(population)
	hasher := murmur3.New128()
	hasher.Write([]byte(person.InternalId))
	partition, _ := hasher.Sum128()
	partition = partition % uint64(numberOfWorkers)
	fmt.Printf("Partition: %d\n", partition)
	text := fmt.Sprintf("--> Starting Enrollment Business Process for person: %s", person.FirstName+" "+person.LastName)
	slog.Debug(text)
	startParameters := businessProcess.BusinessProcessStartContext{
		Person:                     person,
		BusinessProcessDefinition:  businessProcessDefinition,
		EffectiveDate:              effectiveDate,
		SourceEventReferenceNumber: "",
		SourceType:                 "Batch",
	}
	refNum, _ = businessProcess.StartPersonBusinessProcess(ctx, rc, startParameters)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	slog.Debug("********************************")
	slog.Debug("Submitting Enrollment Events")
	slog.Debug("********************************")
	var wg sync.WaitGroup
	ed.StartConsumingMessages(ctx, &wg, rc, sigChan)
	for _, anEvent := range events {
		if anEvent.GetReferenceNumber() == "UNKNOWN" {
			anEvent.SetReferenceNumber(refNum)
		}
		//partitionChannels[int(partition)] <- anEvent
		_ = ed.Publish(ctx, anEvent, int(partition))

	}
	/*
		for i := 0; i < numberOfWorkers; i++ {
			close(partitionChannels[i])

		}
	*/
	wg.Wait()
	ed.Close()
	// Close
	slog.Debug("End Event Distributor")
}

/*
	func processEventsOld(events []message.Event) {
		opts := slog.HandlerOptions{
			//Level: slog.LevelInfo,
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
		slog.SetDefault(logger)
		slog.Debug("Starting Event Distributor")
		mongoDB := businessProcess.NewMongo()
		rc := processors.NewResourceContext(mongoDB)
		defer mongoDB.CloseClientDB()

		context := context.Background()
		numberOfWorkers := 1
		ed := processors.NewChannelEventBroker(mongoDB, numberOfWorkers)

		//eventChannel := make(chan anEvent.Event, 10)
		partitionChannels := make(map[int]chan message.Event, numberOfWorkers)
		for i := 0; i < numberOfWorkers; i++ {
			partitionChannel := make(chan message.Event, numberOfWorkers)
			partitionChannels[i] = partitionChannel
		}

		population, _ := mongoDB.GetPersons()
		effectiveDate := businessProcess.YYYYMMDD_Date("20240101")
		benefitPlanDefinition := businessProcess.CreateMockBusinessProcessDefinitionObjects("BP001")
		mongoDB.DeleteAllPersonBusinessProcesses()
		var refNum string
		slog.Debug("********************************")
		slog.Debug("Starting Enrollment Event")
		slog.Debug("********************************")
		person := getFirstPerson(population)
		hasher := murmur3.New128()
		hasher.Write([]byte(person.InternalId))
		partition, _ := hasher.Sum128()
		partition = partition % uint64(numberOfWorkers)
		fmt.Printf("Partition: %d\n", partition)
		text := fmt.Sprintf("--> Starting Enrollment Business Process for person: %s", person.FirstName+" "+person.LastName)
		slog.Debug(text)
		refNum, _ = businessProcess.StartPersonBusinessProcess(context, rc, person, benefitPlanDefinition, effectiveDate)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		slog.Debug("********************************")
		slog.Debug("Submitting Enrollment Events")
		slog.Debug("********************************")
		var wg *sync.WaitGroup
		ed.StartConsumingEvents(context, wg, rc, sigChan)
		for _, anEvent := range events {
			anEvent.Data.ReferenceNumber = refNum
			partitionChannels[int(partition)] <- anEvent
		}

		for i := 0; i < numberOfWorkers; i++ {
			close(partitionChannels[i])

		}

		wg.Wait()
		// Close
		slog.Debug("End Event Distributor")
	}
*/
func Test_KafkaEventBroker_EnrollInBenefit(t *testing.T) {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	slog.Debug("Deleting Event Topic - Message")
	//kafka.TopicSetup()
	slog.Debug("Starting Kafka Event Distributor")
	personMongoDB := db.NewPersonMongo()
	planMongoDB := db.NewPlanMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ed := businessProcess.NewKafkaMessageBroker(personMongoDB, businessProcessMongoDB, nil)
	ev := datatypes.EnvironmentVariables{}
	rc := businessProcess.NewResourceContext("scenarios::Test_KafkaEventBroker_EnrollInBenefit", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)
	defer personMongoDB.CloseClientDB()
	defer planMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	defer ed.Close()
	defer kafka.DeleteMessageTopic()
	ctx := context.TODO()
	businessProcessMongoDB.DeleteAllPersonBusinessProcesses()
	numberOfPartitions := 1
	//Get Person
	slog.Debug("********************************")
	slog.Debug("Starting Enrollment Event")
	slog.Debug("********************************")
	population, _ := personMongoDB.GetPersons()
	person := getFirstPerson(population)
	effectiveDate := datatypes.YYYYMMDD_Date("20240101")
	businessProcessDefinition := bpd.GetBusinessProcessDefinition("BP001")
	var refNum string
	hasher := murmur3.New128()
	hasher.Write([]byte(person.InternalId))
	partition, _ := hasher.Sum128()
	partition = partition % uint64(numberOfPartitions)
	s := fmt.Sprintf("--> Starting Enrollment Business Process for person: %s", person.FirstName+" "+person.LastName)
	slog.Debug(s)
	startParameters := businessProcess.BusinessProcessStartContext{
		Person:                     person,
		BusinessProcessDefinition:  businessProcessDefinition,
		EffectiveDate:              effectiveDate,
		SourceEventReferenceNumber: "",
		SourceType:                 "Batch",
	}
	refNum, _ = businessProcess.StartPersonBusinessProcess(ctx, rc, startParameters)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	ed.StartConsumingMessages(ctx, &wg, rc, sigChan)
	events := buildEnrollmentEventSequence1(refNum)
	slog.Debug("********************************")
	slog.Debug("Submitting Enrollment Events")
	slog.Debug("********************************")
	kafkaPublisher := ed.GetPublisher()
	for _, anEvent := range events {
		jsonEvent, _ := json.Marshal(anEvent)
		s := fmt.Sprintf("--> Publishing anEvent %s for person: %s", anEvent.GetMessageName(), person.FirstName+" "+person.LastName)
		slog.Debug(s)
		slog.Debug(string(jsonEvent))
		ed.Publish(ctx, anEvent, int(partition))
	}

	slog.Debug(" At Wait")
	wg.Wait()
	kafkaPublisher.Close()
	slog.Debug("End Kafka Event Distributor")
}
func getFirstPerson(population []*person.Person) *person.Person {
	return population[0]

}
func buildEnrollmentEventSequence1(refNum string) []message.Message {
	events := make([]message.Message, 0)
	anEvent := message.BuildCommand("PC0001", refNum)

	events = append(events, anEvent)
	anEvent = message.BuildEvent("PE0002", refNum)
	events = append(events, anEvent)
	anEvent = message.BuildEvent("PE0003", refNum)
	events = append(events, anEvent)
	anEvent = message.BuildEvent("PE0004", refNum)
	events = append(events, anEvent)

	return events
}
func buildEnrollmentEventSequence2(refNum string) []message.Message {

	events := make([]message.Message, 0)
	/*
		anEvent := message.BuildEvent("S1-E1", "EnrollInBenefit", refNum)
		events = append(events, anEvent)
		anEvent = message.BuildEvent("S1-E1", "EnrollInBenefit", refNum)
		events = append(events, anEvent)
		anEvent = message.BuildEvent("S4-E1", "Confirmation Statement", refNum)
		events = append(events, anEvent)
		anEvent = message.BuildEvent("S2-E1", "Release Payroll", refNum)
		events = append(events, anEvent)
		anEvent = message.BuildEvent("S3-E1", "Release Carrier", refNum)
		events = append(events, anEvent)
	*/
	return events

}

/*
	anEvent = message.Event{
		ID:              "S1-E1",
		Name:            "EnrollInBenefit",
		ReferenceNumber: refNum,
	}

eventChannel <- anEvent

	anEvent = message.Event{
		ID:              "S2-E1",
		Name:            "Release Payroll",
		ReferenceNumber: refNum,
	}

	anEvent := message.Event{
				ID:              "NewDay",
				Name:            "Today",
				ReferenceNumber: refNum,
				Data_Date:       "2023-09-25",
			}
*/
