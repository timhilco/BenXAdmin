package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func DeleteMessageTopic() {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)

	DeleteTopic()
}
func CreateMessageTopic() {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)

	CreateTopic()
}
func TopicSetup() {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	slog.Info("Deleting Topic")
	DeleteMessageTopic()
	slog.Info("Sleeping 30 seconds ...")
	time.Sleep(30 * time.Second)
	slog.Info("Creating Topic")
	CreateMessageTopic()
}
func DeleteTopic() {
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		os.Exit(1)
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete topics on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("20s")
	if err != nil {
		panic("ParseDuration(10s)")
	}
	topics := make([]string, 0)
	topics = append(topics, "BenXMessage")
	results, err := a.DeleteTopics(ctx, topics, kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to delete topics: %v\n", err)
		os.Exit(1)
	}

	// Print results
	for _, result := range results {
		fmt.Printf(" Delete Topic: %s\n", result)
	}

	a.Close()
}
func CreateTopic() {
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		os.Exit(1)
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete topics on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		panic("ParseDuration(60s)")
	}
	topic := "BenXMessage"
	results, err := a.CreateTopics(
		ctx,
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     10,
			ReplicationFactor: 1}},
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to create topic: %v\n", err)
		os.Exit(1)
	}

	// Print results
	for _, result := range results {
		fmt.Printf("Create Topic: %s\n", result)
	}
	a.Close()

}
func deleteConsumerGroup(group string, config *kafka.ConfigMap) {

	args := os.Args

	if len(args) < 4 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s <bootstrap_servers> <request_timeout_sec> <group1> [<group2> ...]\n", args[0])
		os.Exit(1)
	}

	// Create new AdminClient.
	ac, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		os.Exit(1)
	}
	defer ac.Close()

	timeoutSec := 15
	if err != nil {
		fmt.Printf("Failed to parse timeout: %s\n", err)
		os.Exit(1)
	}

	groups := []string{group}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := ac.DeleteConsumerGroups(ctx, groups,
		kafka.SetAdminRequestTimeout(time.Duration(timeoutSec)*time.Second))
	if err != nil {
		fmt.Printf("Failed to delete groups: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("DeleteConsumerGroups result: %+v\n", res)

}
