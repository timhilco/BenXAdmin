package db

import (
	"benefitsDomain/domain/person/personRoles"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m PersonMongoDB) GetWorkers(config map[string]string) ([]*personRoles.Worker, error) {
	res, err := m.workerCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		slog.Error("Error while fetching workers:" + err.Error())
		return nil, err
	}
	var workers []*personRoles.Worker
	var aPersonId, keyType string
	keyType = config["keyType"]
	aPersonId = config["PersonId"]
	var filter primitive.M
	var results []*personRoles.Worker
	switch keyType {

	case "PersonId":
		slog.Debug("Worker -> Get documents for Person (External): " + aPersonId)
		filter = bson.M{"personId": bson.D{{Key: "$eq", Value: aPersonId}}}

		// Passing bson.D{{}} as the filter matches all documents in the workerCollection
		//filter := bson.D{{}}
		findOptions := options.Find()
		cur, err := m.workerCollection.Find(context.TODO(), filter, findOptions)
		if err != nil {
			slog.Error(fmt.Sprintf("WorkerGet -> Filter Error: %s", err))
		}

		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		for cur.Next(context.TODO()) {

			// create a value into which the single document can be decoded
			var elem personRoles.Worker
			err := cur.Decode(&elem)
			if err != nil {
				slog.Error(fmt.Sprintf("Worker Get -> Decode Error ->%s", err))
			}

			results = append(results, &elem)
		}

		if err := cur.Err(); err != nil {
			slog.Error(fmt.Sprintf("WorkerDB Get -> Cur Error ->%s", err))
		}

		// Close the cursor once finished
		cur.Close(context.TODO())
		if len(results) == 0 {
			return nil, errors.New("worker not found")
		}
		workers = results
	default:
		err = res.All(context.TODO(), &workers)
		if err != nil {
			slog.Error("Error while decoding workers:" + err.Error())
			return nil, err
		}
	}
	return workers, nil
}

// GetWorker gets documents from the Worker Collection
func (m PersonMongoDB) GetWorker(aWorkerID string, keyType string) (*personRoles.Worker, error) {
	// Pass these options to the Find method
	slog.Debug("Worker -> Get document for: " + aWorkerID + " KeyType: " + keyType)
	findOptions := options.Find()
	findOptions.SetLimit(2)

	// Here's an array in which you can store the decoded documents
	var results []*personRoles.Worker

	collection := m.workerCollection
	var filter primitive.M
	// Passing bson.D{{}} as the filter matches all documents in the collection
	switch keyType {
	case "Internal":
		filter = bson.M{"internalId": bson.D{{Key: "$eq", Value: aWorkerID}}}
	case "Person":
		filter = bson.M{"personId": bson.D{{Key: "$eq", Value: aWorkerID}}}
	case "Worker":
		filter = bson.M{"workerId": bson.D{{Key: "$eq", Value: aWorkerID}}}
	}
	//filter := bson.D{{}}

	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		slog.Error(fmt.Sprintf("WorkerGet -> Filter Error: %s", err))
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem personRoles.Worker
		err := cur.Decode(&elem)
		if err != nil {
			slog.Error(fmt.Sprintf("Worker Get -> Decode Error ->%s", err))
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		slog.Error(fmt.Sprintf("WorkerDB Get -> Cur Error ->%s", err))
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	if len(results) == 0 {
		return nil, errors.New("worker not found")
	}
	worker := results[0]
	return worker, nil
}
func (m PersonMongoDB) InsertWorker(worker *personRoles.Worker) error {
	slog.Debug("WorkerDB -> Inserting document for: " + worker.InternalId)

	collection := m.workerCollection

	insertResult, err := collection.InsertOne(context.TODO(), worker)
	if err != nil {
		slog.Error(fmt.Sprintf("WorkerDB -> Error: %s", err))
	}

	slog.Debug(fmt.Sprintln("Inserted a single document: ", insertResult.InsertedID))
	return nil
}
func (m PersonMongoDB) DeleteAllWorkers() error {
	slog.Debug("Worker -> Deleting All documents")

	deleteResult, err := m.workerCollection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		slog.Error(fmt.Sprintf("WorkerDB -> Error: %s", err))
	}
	s := fmt.Sprintf("Deleted %v documents in the Worker collection\n", deleteResult.DeletedCount)
	slog.Debug(s)
	return nil
}

/*
// UpdateWorker inserts documents into the Worker Collection
func (m *PersonMongoDB) UpdateWorker(ctx context.Context, key string, worker *worker.Worker) error {
	p.logger.Debug("WorkerDB -> Updating document for: " + key)
	Worker.WaitingExpectationsID = deconstructWaitingExpecations(Worker)
	filter := bson.M{"internalID": bson.D{{Key: "$eq", Value: key}}}

	collection := p.MongoClient.Database("WorkerDB").Collection("Worker")
	updateResult, err := collection.ReplaceOne(context.TODO(), filter, Worker)
	if err != nil {
		p.logger.Fatal(fmt.Sprintf("WorkerDB -> Error: %s", err))
	}

	s := fmt.Sprintf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	p.logger.Debug(s)
	return nil
}

*/
