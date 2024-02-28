package kafka

import (
	"testing"
)

func Test_Kafka_DeleteTopic(t *testing.T) {

	DeleteMessageTopic()

}
func Test_Kaka_Setup(t *testing.T) {

	TopicSetup()

}
func Test_Kaka_CreateTopic(t *testing.T) {

	CreateMessageTopic()

}

func Test_Kafka2(t *testing.T) {
	/*
		opts := slog.HandlerOptions{
			//Level: slog.LevelInfo,
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
		slog.SetDefault(logger)
		var wg sync.WaitGroup
		//consumerConfig, _ := NewKafkaConfiguration("Consumer")
		producerConfig, _ := NewKafkaConfiguration("Producer")
		topics := []string{"Event"}
		topic := "Event"
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		kec := NewKafkaConsumer(consumerConfig, topics, sigChan)
		kep := NewKafkaProducer(producerConfig, topic)
		wg.Add(1)
		//go kec.ConsumeEvents(&wg, sigChan)
		wg.Add(1)
		kep.PublishEvent(&wg, []byte{}, 0)
		wg.Wait()
	*/
}
