package businessProcess

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"server/message"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaMessageConsumer struct {
	kafkaConsumer *kafka.Consumer
}
type KafkaMessagePublisher struct {
	kafkaProducer *kafka.Producer
	topic         string
}

func NewKafkaConfiguration(configType string, configMap map[string]string) (*kafka.ConfigMap, error) {
	clientId := "BenX"
	groupId := "Hilco1"
	if configMap != nil {
		gi, ok := configMap["group.id"]
		if ok {
			groupId = gi
		}
		ci, ok := configMap["client.id"]
		if ok {
			clientId = ci
		}

	}
	switch configType {
	case "Consumer":
		kc := &kafka.ConfigMap{
			"bootstrap.servers":        "localhost:9092",
			"client.id":                clientId,
			"group.id":                 groupId,
			"broker.address.family":    "v4",
			"session.timeout.ms":       6000,
			"auto.offset.reset":        "earliest",
			"enable.auto.offset.store": false,
		}
		return kc, nil
	case "Producer":
		kc := &kafka.ConfigMap{
			"bootstrap.servers":     "localhost:9092",
			"client.id":             clientId,
			"acks":                  "all",
			"broker.address.family": "v4",
		}
		return kc, nil
	}
	return nil, nil
}
func NewKafkaConsumer(kafkaConfig *kafka.ConfigMap, topics []string, sigchan chan os.Signal) KafkaMessageConsumer {
	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		slog.Error("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}
	s := fmt.Sprintf("Consumer - Created Consumer: %v", c)
	slog.Info(s)

	err = c.SubscribeTopics(topics, nil)

	if err != nil {
		slog.Error("Consumer - Failed to create consumer: %s\n", err)
		os.Exit(1)
	}
	return KafkaMessageConsumer{c}
}

func NewKafkaProducer(kafkaConfig *kafka.ConfigMap, topic string) KafkaMessagePublisher {
	p, err := kafka.NewProducer(kafkaConfig)

	if err != nil {
		slog.Error("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	s := fmt.Sprintf("Created Producer: %v\n", p)
	slog.Info(s)

	// Listen to all the events on the default events channel
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				// The message delivery report, indicating success or
				// permanent failure after retries have been exhausted.
				// Application level retries won't help since the client
				// is already configured to do that.
				m := ev
				if m.TopicPartition.Error != nil {
					s := fmt.Sprintf("Producer -Delivery failed: %v", m.TopicPartition.Error)
					slog.Info(s)
				} else {
					s := fmt.Sprintf("Producer -Delivered message to topic %s [%d] at offset %v",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
					slog.Info(s)
				}
			case kafka.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				//
				// These errors should generally be considered informational
				// as the underlying client will automatically try to
				// recover from any errors encountered, the application
				// does not need to take action on them.
				s := fmt.Sprintf("Producer -Error: %v\n", ev)
				slog.Info(s)
			default:
				s := fmt.Sprintf("Producer - Ignored event: %s\n", ev)
				slog.Info(s)
			}
		}
	}()
	return KafkaMessagePublisher{p,
		topic}
}
func (p KafkaMessageConsumer) Close() {
	p.kafkaConsumer.Close()
}

func (p KafkaMessagePublisher) PublishMessage(aMessage message.Message, partition int32) {
	topic := p.topic
	value := fmt.Sprintf("Producer message: %s", aMessage)
	slog.Debug(value)
	var bMessage []byte
	var headerValue string

	headerKey := "Type"
	switch aMessage.(type) {
	case *message.Event:
		event, _ := aMessage.MarshallJson()
		bMessage = event
		headerValue = message.C_MESSAGE_TYPE_PERSON_EVENT
	case *message.Command:
		command, _ := aMessage.MarshallJson()
		bMessage = command
		headerValue = message.C_MESSAGE_TYPE_COMMAND

	}

	err := p.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: partition},
		Value:          bMessage,
		Headers:        []kafka.Header{{Key: headerKey, Value: []byte(headerValue)}},
	}, nil)

	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrQueueFull {
			// Producer queue is full, wait 1s for messages
			// to be delivered then try again.
			time.Sleep(time.Second)
			//continue
		}
		s := fmt.Sprintf("Producer - Failed to produce message: %v\n", err)
		slog.Info(s)
	}

	// Flush and close the producer and the events channel
	for p.kafkaProducer.Flush(10000) > 0 {
		fmt.Print("Producer - Still waiting to flush outstanding messages\n")
	}

}
func (p KafkaMessagePublisher) Close() {
	p.kafkaProducer.Close()
}
func (c KafkaMessageConsumer) ConsumeMessages(ctx context.Context, wg *sync.WaitGroup, sigchan chan os.Signal, f func(context.Context, *ResourceContext, message.Message, int) error, rc *ResourceContext) {

	defer wg.Done()
	run := true

	for run {
		select {
		case sig := <-sigchan:
			s := fmt.Sprintf("ConsumeMessages -Caught signal %v: terminating\n", sig)
			slog.Info(s)
			run = false
		default:
			ev := c.kafkaConsumer.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				// Process the message received.
				slog.Info("Consumer *kafka.Message case entered")
				s := fmt.Sprintf("%% ConsumeMessages - Message on %s",
					e.TopicPartition)
				slog.Info(s)
				s = fmt.Sprintf("%% ConsumeMessages - Message Value:%s",
					string(e.Value))
				slog.Debug(s)
				if e.Headers != nil {
					s := fmt.Sprintf("%% ConsumeMessages - Headers: %v", e.Headers)
					slog.Debug(s)
				}
				typeHeader := e.Headers[0]
				var aMessage message.Message
				messageType := string(typeHeader.Value)
				switch messageType {
				case message.C_MESSAGE_TYPE_PERSON_EVENT:
					var event message.Event
					err := json.Unmarshal(e.Value, &event)
					if err != nil {
						s := fmt.Sprintf("%% ConsumeMessages - Event Message Type: Error:  %v\n", err)
						slog.Error(s)
					}
					aMessage = &event
				case message.C_MESSAGE_TYPE_COMMAND:
					var command message.Command
					err := json.Unmarshal(e.Value, &command)
					if err != nil {
						s := fmt.Sprintf("%%ConsumeMessages - Command Message Type: Error:  %v\n", err)
						slog.Error(s)
					}
					aMessage = &command

				}
				eId := aMessage.GetMessageId()
				s = fmt.Sprintf("ConsumeMessages - Calling Distribution function for EventId: %s", eId)
				slog.Debug(s)
				f(ctx, rc, aMessage, 0)
				// We can store the offsets of the messages manually or let
				// the library do it automatically based on the setting
				// enable.auto.offset.store. Once an offset is stored, the
				// library takes care of periodically committing it to the broker
				// if enable.auto.commit isn't set to false (the default is true).
				// By storing the offsets manually after completely processing
				// each message, we can ensure atleast once processing.
				_, err := c.kafkaConsumer.StoreMessage(e)
				if err != nil {
					s := fmt.Sprintf("%% Error storing offset after message %s:\n",
						e.TopicPartition)
					slog.Info(s)
				}
				slog.Info("---------------------------")
			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				slog.Info("$$$$$$$$$$$$$$$$$$$$$$$$$$$")
				slog.Info("Consumer *kafka.Error case entered")
				// But in this example we choose to terminate
				// the application if all brokers are down.
				s := fmt.Sprintf("%% Error: %v: %v\n", e.Code(), e)
				slog.Info(s)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
				slog.Info("$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			case kafka.OffsetsCommitted:
				slog.Info("$$$$$$$$$$$$$$$$$$$$$$$$$$$")
				slog.Info("Consumer *kafka.OffsetCommitted case entered")
				if e.Error != nil {
					slog.Info("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
					s := fmt.Sprintf("**** Error: %v\n", e.Error)
					slog.Info("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
					slog.Info(s)
				}
				for _, tp := range e.Offsets {
					s := fmt.Sprintf("%% --TopicPartition: %v", tp)
					slog.Info(s)
				}
				slog.Info("$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			case kafka.AssignedPartitions:
				slog.Info("===========================")
				slog.Info("Consumer *kafka.AssignedPartition case entered")
				for _, tp := range e.Partitions {
					s := fmt.Sprintf("%% --TopicPartition: %v", tp)
					slog.Info(s)
				}
				slog.Info("===========================")
			case kafka.RevokedPartitions:
				slog.Info("===========================")
				slog.Info("Consumer *kafka.RevokedPartition case entered")
				for _, tp := range e.Partitions {
					s := fmt.Sprintf("%% --TopicPartition: %v", tp)
					slog.Info(s)
				}
				slog.Info("===========================")
			case kafka.PartitionEOF:
				slog.Info("===========================")
				slog.Info("Consumer *kafka.PartitionEOF case entered")
				slog.Info(e.String())
				slog.Info("===========================")
			default:
				slog.Info("$$$$$$$$$$$$$$$$$$$$$$$$$$$")
				s := fmt.Sprintf("Consumer default case entered - Ignored %v\n", e)
				slog.Info(s)
				slog.Info("$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			}
		}
	}

	slog.Debug("Closing consumer")
	c.kafkaConsumer.Close()

}
