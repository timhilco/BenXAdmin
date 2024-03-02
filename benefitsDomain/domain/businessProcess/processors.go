package businessProcess

import (
	"benefitsDomain/domain/db"
	"benefitsDomain/domain/person"
	"benefitsDomain/domain/person/personRoles"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"server/message"
	"strconv"
	"sync"
	"syscall"
	"text/template"
	"time"
)

type MessageProcessingContext struct {
	ResourceContext                   *ResourceContext
	Message                           message.Message
	Person                            *person.Person
	Worker                            *personRoles.Worker
	PersonBusinessProcess             *PersonBusinessProcess
	BusinessProcessDefinition         *BusinessProcessDefinition
	ShouldUpdatePersonBusinessProcess bool
	ContextDataMap                    map[string]string
	ProcessingArrayIndex              int
}

func (m *MessageProcessingContext) String() string {
	return m.Report(nil)
}
func (m *MessageProcessingContext) GetMessage() message.Message {
	return m.Message
}
func (d *MessageProcessingContext) Report(rc *ResourceContext) string {
	dir := rc.environmentVariables.TemplateDirectory
	templateFile := dir + "messageProcessingContextTemplate.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, d)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type MessageBroker interface {
	DistributeMessage(context.Context, *ResourceContext, message.Message, int) error
	StartConsumingMessages(ctx context.Context, wg *sync.WaitGroup, rc *ResourceContext, sigchan chan os.Signal)
	Close() error
	Publish(context.Context, message.Message, int) error
	GetNumberOfPartitions() int
}

type ChannelMessageBroker struct {
	personDomainStore          *db.PersonMongoDB
	businessProcessDomainStore *BusinessProcessMongoDB
	//consumer  *KafkaMessageConsumer
	partitionChannels map[int]chan message.Message
	//numberOfPartition int
}
type KafkaMessageBroker struct {
	brokerType                 string
	personDomainStore          *db.PersonMongoDB
	businessProcessDomainStore *BusinessProcessMongoDB
	consumer                   *KafkaMessageConsumer
	publisher                  *KafkaMessagePublisher
	numberOfPartition          int
}

func NewChannelMessageBroker(personMongoDB *db.PersonMongoDB, businessProcessMongoDb *BusinessProcessMongoDB, partitions int) MessageBroker {
	partitionChannels := make(map[int]chan message.Message, partitions)
	for i := 0; i < partitions; i++ {
		partitionChannel := make(chan message.Message, 100)
		partitionChannels[i] = partitionChannel
	}
	ed := &ChannelMessageBroker{
		personDomainStore:          personMongoDB,
		businessProcessDomainStore: businessProcessMongoDb,
		partitionChannels:          partitionChannels,
	}
	return ed
}
func (ceb *ChannelMessageBroker) StartConsumingMessages(ctx context.Context, wg *sync.WaitGroup, rc *ResourceContext, sigchan chan os.Signal) {
	for i := 0; i < len(ceb.partitionChannels); i++ {
		wg.Add(1)
		go processMessage(ceb.partitionChannels[i], ctx, rc, ceb, i, wg, sigchan)
	}
}

func processMessage(jobChannel chan message.Message, context context.Context, rc *ResourceContext, ed MessageBroker, workerId int, wg *sync.WaitGroup, sigchan chan os.Signal) {
	defer wg.Done()

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case anMessage := <-jobChannel:
			messageId := anMessage.GetMessageId()
			if messageId == "CloseEventPartition" {
				slog.Debug("processMessage: Close Event Partition - Quitting")
				return
			} else {
				slog.Debug("processMessage: Received Message: " + anMessage.String())
				err := ed.DistributeMessage(context, rc, anMessage, workerId)
				if err != nil {
					slog.Error("Error", err)
				}
			}
		case sig := <-sigchan:
			slog.Debug("processMessage: sigchan - Quitting:  " + sig.String())
			return
		case <-ticker.C:
			slog.Debug("processMessage: Ticker - Waiting for message")

		}
	}
}
func (ceb *ChannelMessageBroker) Publish(ctx context.Context, e message.Message, partition int) error {
	sp := " -> Partition: " + strconv.Itoa(partition)
	slog.Debug("ChannelMessageBroker::Publish : " + e.String() + sp)

	ch := ceb.partitionChannels[partition]
	ch <- e
	return nil
}
func (ceb *ChannelMessageBroker) Close() error {
	for i, c := range ceb.partitionChannels {
		slog.Debug("ChannelMessageBroker::Close : " + strconv.Itoa(i))
		close(c)

	}
	return nil

}

const (
	C_THIS = iota
	C_INTERESTED_PARTIES
)

func (e *ChannelMessageBroker) DistributeMessage(ctx context.Context, rc *ResourceContext, event message.Message, workerId int) error {
	fmt.Println(event.Report("./templates/"))
	refnum := event.GetReferenceNumber()
	target := event.GetTarget()
	var action int
	pbp, _ := rc.GetBusinessProcessStore().GetPersonBusinessProcess(refnum)
	if target == "This" {
		action = C_THIS
	} else {
		action = C_INTERESTED_PARTIES
	}
	parties := make([]*PersonBusinessProcess, 0)

	switch action {
	case C_THIS:
		parties = append(parties, pbp)
	case C_INTERESTED_PARTIES:
		parties, _ = findInterestedParties(ctx, rc, event, refnum)
	}
	if len(parties) != 0 {
		for _, p := range parties {
			person, _ := rc.GetPersonDataStore().GetPerson(p.PersonId, "Internal")
			s := fmt.Sprintf("Person: %s - Message: %s - Business Process: %s ", person.LastName, event.GetMessageName(), p.ReferenceNumber)
			slog.Debug("--> Starting Message Distribution : " + s)
			if shouldProcess(p, event) {
				p.ProcessMessage(rc, event)
			} else {
				slog.Info("--> Message not processed -ShouldProcess")
			}
		}
	} else {
		slog.Info("--> Message not processed - Parties = 0")
	}

	slog.Debug("Ending Distribution of event: ")
	return nil
}
func (ced *ChannelMessageBroker) GetNumberOfPartitions() int {
	return 10
}

func findInterestedParties(ctx context.Context, rc *ResourceContext, event message.Message, refNum string) ([]*PersonBusinessProcess, error) {
	interestedParties := make([]*PersonBusinessProcess, 0)
	var config map[string]string
	pbp, _ := rc.GetBusinessProcessStore().GetPersonBusinessProcesses(config)
	for _, p := range pbp {
		if p.SourceEventReferenceNumber == refNum {
			interestedParties = append(interestedParties, p)

		}
	}
	return interestedParties, nil
}
func shouldProcess(p *PersonBusinessProcess, m message.Message) bool {
	return p.State != C_STATE_CLOSED

}
func NewKafkaMessageBroker(personMongoDB *db.PersonMongoDB, businessProcessMongoDb *BusinessProcessMongoDB, configMap map[string]string) *KafkaMessageBroker {

	consumerConfig, _ := NewKafkaConfiguration("Consumer", configMap)
	producerConfig, _ := NewKafkaConfiguration("Producer", configMap)
	topics := []string{"BenXMessage"}
	topic := "BenXMessage"
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	kec := NewKafkaConsumer(consumerConfig, topics, sigChan)
	kep := NewKafkaProducer(producerConfig, topic)
	ed := &KafkaMessageBroker{
		personDomainStore:          personMongoDB,
		businessProcessDomainStore: businessProcessMongoDb,
		consumer:                   &kec,
		publisher:                  &kep,
		numberOfPartition:          10,
	}
	return ed
}
func NewKafkaConsumerMessageBroker(personMongoDB *db.PersonMongoDB, businessProcessMongoDb *BusinessProcessMongoDB, configMap map[string]string) *KafkaMessageBroker {
	consumerConfig, _ := NewKafkaConfiguration("Consumer", configMap)
	topics := []string{"Message"}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	kec := NewKafkaConsumer(consumerConfig, topics, sigChan)
	//topic := "Message"
	//producerConfig, _ := NewKafkaConfiguration("Producer")
	//kep := NewKafkaProducer(producerConfig, topic)
	ed := &KafkaMessageBroker{
		brokerType:                 "Consumer",
		personDomainStore:          personMongoDB,
		businessProcessDomainStore: businessProcessMongoDb,
		consumer:                   &kec,
		publisher:                  nil,
		numberOfPartition:          10,
	}
	return ed
}
func NewKafkaProducerMessageBroker(personMongoDB *db.PersonMongoDB, businessProcessMongoDb *BusinessProcessMongoDB, configMap map[string]string) *KafkaMessageBroker {
	producerConfig, _ := NewKafkaConfiguration("Producer", nil)
	topic := "BenXMessage"
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	//topics := []string{"Message"}
	//consumerConfig, _ := NewKafkaConfiguration("Consumer")
	//kec := NewKafkaConsumer(consumerConfig, topics, sigChan)
	kep := NewKafkaProducer(producerConfig, topic)
	ed := &KafkaMessageBroker{
		brokerType:                 "Producer",
		personDomainStore:          personMongoDB,
		businessProcessDomainStore: businessProcessMongoDb,
		consumer:                   nil,
		publisher:                  &kep,
		numberOfPartition:          10,
	}
	return ed
}

func (c *KafkaMessageBroker) GetConsumer() *KafkaMessageConsumer {
	return c.consumer

}
func (c *KafkaMessageBroker) GetPublisher() *KafkaMessagePublisher {
	return c.publisher

}
func (c *KafkaMessageBroker) ClosePublisher() {
	slog.Info("Closing Kafka Publisher")
	kep := c.publisher
	kep.Close()

}
func (c *KafkaMessageBroker) CloseConsumer() {
	kec := c.consumer
	if kec != nil {
		slog.Info("Closing Kafka Consumer")
		kec.Close()
	} else {
		slog.Info("Kafka Consumer already closed")

	}

}
func (ked *KafkaMessageBroker) StartConsumingMessages(ctx context.Context, wg *sync.WaitGroup, rc *ResourceContext, sigchan chan os.Signal) {
	wg.Add(1)
	consumer := ked.GetConsumer()
	f := ked.DistributeMessage
	go consumer.ConsumeMessages(ctx, wg, sigchan, f, rc)

}
func (ked *KafkaMessageBroker) DistributeMessage(ctx context.Context, rc *ResourceContext, event message.Message, worker int) error {
	messageId := event.GetMessageId()
	switch messageId {
	case "CloseEventPartition":
		slog.Debug("processMessage: Close Event Partition - Ignore for Kafka")
	default:

		refnum := event.GetReferenceNumber()
		personBusinessProcess, err := rc.GetBusinessProcessStore().GetPersonBusinessProcess(refnum)
		if err != nil {
			slog.Info("DistributeMessage Error: Person Business Process Object not found " + refnum)
		} else {
			person, _ := rc.GetPersonDataStore().GetPerson(personBusinessProcess.PersonId, "Internal")
			s := fmt.Sprintf("Person: %s - Message: %s - Business Process: %s  ", person.LastName, event.GetMessageName(), personBusinessProcess.ReferenceNumber)
			slog.Info("--> Starting MessageDistribution : " + s)
			personBusinessProcess.ProcessMessage(rc, event)
			slog.Info("Ending Distribution of event: " + s)
		}
	}
	return nil
}
func (ked *KafkaMessageBroker) GetNumberOfPartitions() int {
	return ked.numberOfPartition
}

func (ked *KafkaMessageBroker) Close() error {
	ked.ClosePublisher()
	ked.CloseConsumer()
	return nil
}
func (keb *KafkaMessageBroker) Publish(ctx context.Context, e message.Message, partition int) error {
	n := keb.GetNumberOfPartitions()
	if partition == -1 || partition > n {
		partition = partition % n
	}
	slog.Debug("KafkaMessageBroker::Publish : " + e.String())
	publisher := keb.GetPublisher()
	publisher.PublishMessage(e, int32(partition))

	return nil
}
