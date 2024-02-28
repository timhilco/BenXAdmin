package db

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
)

type PlanMongoDB struct {
	client            *mongo.Client
	benefitCollection *mongo.Collection
	// collections if wanting to cache
}

func NewPlanMongo() *PlanMongoDB {
	m := PlanMongoDB{}
	m.client = m.ResolveClientDB()
	m.benefitCollection = m.client.Database("benefit").Collection("benefit")

	return &m
}

func (m *PlanMongoDB) ResolveClientDB() *mongo.Client {
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

func (m *PlanMongoDB) CloseClientDB() {
	if m.client == nil {
		return
	}

	err := m.client.Disconnect(context.TODO())
	if err != nil {
		slog.Error(err.Error())
	}

	fmt.Println("Connection to PlanMongoDB closed.")
}
