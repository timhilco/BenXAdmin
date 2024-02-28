package businessProcess

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BusinessProcessMongoDB struct {
	client                          *mongo.Client
	personBusinessProcessCollection *mongo.Collection
	// collections if wanting to cache
}

func NewBusinessProcessMongo() *BusinessProcessMongoDB {
	m := BusinessProcessMongoDB{}
	m.client = m.ResolveClientDB()
	m.personBusinessProcessCollection = m.client.Database("personBusinessProcess").Collection("personBusinessProcess")
	return &m
}

func GetGlobalInternalIdentifier() string {
	uuid, _ := uuid.NewRandom()
	return uuid.String()

}

func ClientOptions() *options.ClientOptions {
	host := "db"
	if os.Getenv("profile") != "prod" {
		host = "localhost"
	}
	return options.Client().ApplyURI(
		"mongodb://" + host + ":27017",
	)
}

func (m *BusinessProcessMongoDB) ResolveClientDB() *mongo.Client {
	if m.client != nil {
		return m.client
	}
	var err error
	client, err := mongo.Connect(context.TODO(), ClientOptions())
	if err != nil {
		slog.Error(err.Error())
	}
	// check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		slog.Error(err.Error())
	}

	m.client = client
	return client
}

func (m *BusinessProcessMongoDB) CloseClientDB() {
	if m.client == nil {
		return
	}

	err := m.client.Disconnect(context.TODO())
	if err != nil {
		slog.Error(err.Error())
	}

	fmt.Println("Connection to BusinessProcessPersonMongoDB closed.")
}
