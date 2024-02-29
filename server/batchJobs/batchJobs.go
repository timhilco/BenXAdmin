package batchjobs

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"context"
	"fmt"
	"server/kafka"
	"server/message"

	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/spaolacci/murmur3"
)

func StartEnrollmentJob(effectiveDate datatypes.YYYYMMDD_Date, messageBroker string, ev datatypes.EnvironmentVariables) string {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
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
	numberOfWorkers := 1
	var ed businessProcess.MessageBroker
	if messageBroker == "Kafka" {
		ed = businessProcess.NewKafkaMessageBroker(personMongoDB, businessProcessMongoDB, nil)
	} else {
		ed = businessProcess.NewChannelMessageBroker(personMongoDB, businessProcessMongoDB, numberOfWorkers)
	}
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	rc := businessProcess.NewResourceContext("batchjobs::StartEnrollment", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)
	population, _ := personMongoDB.GetPersons()
	businessProcessDefinition := bpd.GetBusinessProcessDefinition("BP001")
	//businessProcessDefinition.Report()
	businessProcessMongoDB.DeleteAllPersonBusinessProcesses()
	personMongoDB.DeleteAllParticipants()
	slog.Debug("********************************")
	slog.Debug("Starting Enrollment Event")
	slog.Debug("********************************")
	for _, person := range population {
		worker, _ := personMongoDB.GetWorker(person.InternalId, "Person")
		startParameters := businessProcess.BusinessProcessStartContext{
			Person:                     person,
			Worker:                     worker,
			BusinessProcessDefinition:  businessProcessDefinition,
			EffectiveDate:              effectiveDate,
			SourceEventReferenceNumber: "",
			SourceType:                 "Batch",
		}
		_, _ = businessProcess.StartPersonBusinessProcess(ctx, rc, startParameters)
	}
	return "Success"
}
func StartKafkaMessageConsumer(personMongoDB *db.PersonMongoDB,
	businessProcessMongoDB *businessProcess.BusinessProcessMongoDB,
	configMap map[string]string,
	bpd businessProcess.BusinessProcessDefinitionDataStore,
	plan *db.PlanMongoDB,
	ev datatypes.EnvironmentVariables) *businessProcess.KafkaMessageBroker {
	slog.Info("Starting Kafka Message Consumer")
	ed := businessProcess.NewKafkaMessageBroker(personMongoDB, businessProcessMongoDB, configMap)
	rcConsumer := businessProcess.NewResourceContext("batchJobs::StartKafkaMessageConsumer", personMongoDB, businessProcessMongoDB, bpd, plan, ed, ev)
	context := context.TODO()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	var wg sync.WaitGroup
	ed.StartConsumingMessages(context, &wg, rcConsumer, sigChan)
	slog.Info("Kafka Message Consumer Started")
	return ed
}
func StartKafkaEnrollmentJob(effectiveDate datatypes.YYYYMMDD_Date) {
	kafka.TopicSetup()
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
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
	context := context.TODO()
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ev := datatypes.EnvironmentVariables{}
	configMap := make(map[string]string)
	configMap["client.id"] = "BenX"
	configMap["group.id"] = "Hilco1"
	StartKafkaMessageConsumer(personMongoDB, businessProcessMongoDB, configMap, bpd, planMongoDB, ev)

	pb := businessProcess.NewKafkaProducerMessageBroker(personMongoDB, businessProcessMongoDB, configMap)
	rcPublisher := businessProcess.NewResourceContext("batchJobs::StartKafkaEnrollmentJob", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, pb, ev)
	population, _ := personMongoDB.GetPersons()
	businessProcessDefinition := bpd.GetBusinessProcessDefinition("BP001")
	//businessProcessDefinition.Report()
	businessProcessMongoDB.DeleteAllPersonBusinessProcesses()
	personMongoDB.DeleteAllParticipants()
	slog.Debug("********************************")
	slog.Debug("Starting Enrollment Event")
	slog.Debug("********************************")
	for _, person := range population {
		worker, _ := personMongoDB.GetWorker(person.InternalId, "Person")
		startParameters := businessProcess.BusinessProcessStartContext{
			Person:                     person,
			Worker:                     worker,
			BusinessProcessDefinition:  businessProcessDefinition,
			EffectiveDate:              effectiveDate,
			SourceEventReferenceNumber: "",
			SourceType:                 "Batch",
		}
		_, _ = businessProcess.StartPersonBusinessProcess(context, rcPublisher, startParameters)
	}
	/*
		hasher := murmur3.New128()
		numberOfPartitions := 10
			businessProcessDefinitionId := "BP001"
			eventIds := [3]string{"PE0002", "PE0003", "PE0004"}
			batch, _ := businessProcessMongoDB.GetPersonBusinessProcesses()
			for _, pbp := range batch {
				if pbp.BusinessProcessDefinitionId == businessProcessDefinitionId {
					hasher.Write([]byte(pbp.InternalId))
					partition, _ := hasher.Sum128()
					partition = partition % uint64(numberOfPartitions)
					fmt.Printf("Partition: %d\n", partition)
					for _, eventId := range eventIds {
						anEvent := message.BuildEvent(eventId, pbp.ReferenceNumber)
						_ = pb.Publish(anEvent, int(partition))
					}

				}

			}
	*/
}
func SendElectionCommand(eventId string, messageBroker string) {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
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
	numberOfWorkers := 10
	var ed businessProcess.MessageBroker
	if messageBroker == "Kafka" {
		ed = businessProcess.NewKafkaMessageBroker(personMongoDB, businessProcessMongoDB, nil)
	} else {
		ed = businessProcess.NewChannelMessageBroker(personMongoDB, businessProcessMongoDB, numberOfWorkers)
	}
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ev := datatypes.EnvironmentVariables{}
	rc := businessProcess.NewResourceContext("batchJobs::SendPopulationEvent", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)
	var config map[string]string
	population, _ := businessProcessMongoDB.GetPersonBusinessProcesses(config)
	businessProcessDefinitionId := "BP001"
	slog.Debug("********************************")
	slog.Debug("Starting Enrollment Population Event")
	slog.Debug("********************************")

	hasher := murmur3.New128()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	slog.Debug("********************************")
	slog.Debug("Submitting Enrollment Events")
	slog.Debug("********************************")
	var wg sync.WaitGroup
	ed.StartConsumingMessages(ctx, &wg, rc, sigChan)
	for _, person := range population {
		if person.BusinessProcessDefinitionId == businessProcessDefinitionId {
			hasher.Write([]byte(person.InternalId))
			partition, _ := hasher.Sum128()
			partition = partition % uint64(numberOfWorkers)
			fmt.Printf("Partition: %d\n", partition)
			anEvent := message.BuildCommand(eventId, person.ReferenceNumber)
			_ = ed.Publish(ctx, anEvent, int(partition))

		}
	}

	anEvent := message.BuildClosePartitionEvent()

	if messageBroker != "Kafka" {
		for i := 0; i < numberOfWorkers; i++ {
			_ = ed.Publish(ctx, anEvent, i)

		}
	}
	wg.Wait()
	ed.Close()
}
func SendPopulationEvent(eventId string, messageBroker string) {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
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
	numberOfWorkers := 1
	var ed businessProcess.MessageBroker
	if messageBroker == "Kafka" {
		ed = businessProcess.NewKafkaMessageBroker(personMongoDB, businessProcessMongoDB, nil)
	} else {
		ed = businessProcess.NewChannelMessageBroker(personMongoDB, businessProcessMongoDB, numberOfWorkers)
	}
	bpd := businessProcess.NewBusinessProcessDefinitionMock()
	ev := datatypes.EnvironmentVariables{}
	rc := businessProcess.NewResourceContext("batchJobs::SendPopulationEvent", personMongoDB, businessProcessMongoDB, bpd, planMongoDB, ed, ev)
	population, _ := businessProcessMongoDB.GetPersonBusinessProcesses(nil)
	businessProcessDefinitionId := "BP001"
	slog.Debug("********************************")
	slog.Debug("Starting Enrollment Population Event")
	slog.Debug("********************************")

	hasher := murmur3.New128()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	slog.Debug("********************************")
	slog.Debug("Submitting Enrollment Events")
	slog.Debug("********************************")
	var wg sync.WaitGroup
	ed.StartConsumingMessages(ctx, &wg, rc, sigChan)
	for _, person := range population {
		if person.BusinessProcessDefinitionId == businessProcessDefinitionId {
			hasher.Write([]byte(person.InternalId))
			partition, _ := hasher.Sum128()
			partition = partition % uint64(numberOfWorkers)
			fmt.Printf("Partition: %d\n", partition)
			anEvent := message.BuildEvent(eventId, person.ReferenceNumber)
			_ = ed.Publish(ctx, anEvent, int(partition))

		}
	}

	anEvent := message.BuildClosePartitionEvent()
	for i := 0; i < numberOfWorkers; i++ {
		_ = ed.Publish(ctx, anEvent, i)

	}
	wg.Wait()
	ed.Close()
}
