package scenarios

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"context"
	"fmt"

	batchjobs "server/batchJobs"
	"server/kafka"
	"server/message"

	"log/slog"
	"os"
	"os/signal"
	"sync"
	"testing"

	"github.com/spaolacci/murmur3"
)

func Test_Enrollment_Start_Population(t *testing.T) {
	effectiveDate := datatypes.YYYYMMDD_Date("20240101")
	ev := datatypes.EnvironmentVariables{}
	batchjobs.StartEnrollmentJob(effectiveDate, "X", ev)
}
func Test_Kafka_Enrollment_Start_Population(t *testing.T) {
	effectiveDate := datatypes.YYYYMMDD_Date("20240101")
	batchjobs.StartKafkaEnrollmentJob(effectiveDate)
}
func Test_Kafka_Elections_Command(t *testing.T) {
	batchjobs.SendElectionCommand("PC0001", "Kafka")
}
func Test_Enrollment_Post_Election_Processes(t *testing.T) {
	batchjobs.SendPopulationEvent("PE0002", "Kafka")
	batchjobs.SendPopulationEvent("PE0003", "Kafka")
	batchjobs.SendPopulationEvent("PE0004", "Kafka")
}
func Test_Person_Complete_Enrollment(t *testing.T) {
	kafka.TopicSetup()
	max := 11

	ExecuteCompleteEnrollment(max, "Kafka")

}
func ExecuteCompleteEnrollment(max int, messageBroker string) error {
	opts := slog.HandlerOptions{
		Level: slog.LevelInfo,
		//Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	slog.Debug("Starting Event Distributor")
	personMongoDB := db.NewPersonMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	planMongoDB := db.NewPlanMongo()
	defer personMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	defer planMongoDB.CloseClientDB()
	ctx := context.TODO()
	numberOfPartitions := 10
	var ed businessProcess.MessageBroker
	if messageBroker == "Kafka" {
		ed = businessProcess.NewKafkaMessageBroker(personMongoDB, businessProcessMongoDB, nil)
	} else {
		ed = businessProcess.NewChannelMessageBroker(personMongoDB, businessProcessMongoDB, numberOfPartitions)
	}
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ev := datatypes.EnvironmentVariables{}
	rc := businessProcess.NewResourceContext("batchJobs::SendPopulationEvent", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)
	hasher := murmur3.New128()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	var wg sync.WaitGroup
	ed.StartConsumingMessages(ctx, &wg, rc, sigChan)
	effectiveDate := datatypes.YYYYMMDD_Date("20240101")
	businessProcessDefinitionId := "BP001"
	businessProcessDefinition := bpd.GetBusinessProcessDefinition("BP001")
	businessProcessMongoDB.DeleteAllPersonBusinessProcesses()
	personMongoDB.DeleteAllParticipants()

	for i := 10; i < max; i++ {
		personId := fmt.Sprintf("0%d-%d-00%d", i, i, i)
		referenceNumber := personId + "_" + businessProcessDefinitionId + "_" + effectiveDate.String()
		person, _ := personMongoDB.GetPerson(personId, "External")
		worker, _ := personMongoDB.GetWorker(person.InternalId, "Person")
		startParameters := businessProcess.BusinessProcessStartContext{
			Person:                     person,
			Worker:                     worker,
			BusinessProcessDefinition:  businessProcessDefinition,
			EffectiveDate:              effectiveDate,
			SourceEventReferenceNumber: "",
			SourceType:                 "Batch",
		}
		slog.Debug("Start Enrollment Business Process")
		_, _ = businessProcess.StartPersonBusinessProcess(ctx, rc, startParameters)

		pbp, _ := businessProcessMongoDB.GetPersonBusinessProcess(referenceNumber)
		slog.Debug("Submit Elections Command")
		anEvent := message.BuildCommand("PC0001", referenceNumber)
		hasher.Write([]byte(person.InternalId))
		partition, _ := hasher.Sum128()
		partition = partition % uint64(numberOfPartitions)
		fmt.Printf("Partition: %d\n", partition)
		_ = ed.Publish(ctx, anEvent, int(partition))
		eventIds := [3]string{"PE0002", "PE0003", "PE0004"}
		for _, eventId := range eventIds {
			slog.Debug("Process Event: " + eventId)
			if pbp.BusinessProcessDefinitionId == businessProcessDefinitionId {
				anEvent := message.BuildEvent(eventId, referenceNumber)
				_ = ed.Publish(ctx, anEvent, int(partition))

			}
		}
	}
	if messageBroker != "Kafka" {
		anEvent := message.BuildClosePartitionEvent()
		for i := 0; i < numberOfPartitions; i++ {
			_ = ed.Publish(ctx, anEvent, i)

		}
	}

	wg.Wait()
	ed.Close()
	return nil
}
