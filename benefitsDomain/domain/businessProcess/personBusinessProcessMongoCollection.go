package businessProcess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m BusinessProcessMongoDB) GetPersonBusinessProcesses() ([]*PersonBusinessProcess, error) {
	res, err := m.personBusinessProcessCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		slog.Error("Error while fetching personBusinessProcess:" + err.Error())
		return nil, err
	}
	var personBusinessProcess []*PersonBusinessProcess
	err = res.All(context.TODO(), &personBusinessProcess)
	if err != nil {
		slog.Error("Error while decoding personBusinessProcess:" + err.Error())
		return nil, err
	}
	return personBusinessProcess, nil
}

// GetPersonBusinessProcess gets documents from the PersonBusinessProcess Collection
func (m BusinessProcessMongoDB) GetPersonBusinessProcess(aPersonBusinessProcessID string) (*PersonBusinessProcess, error) {
	// Pass these options to the Find method
	slog.Debug("PersonBusinessProcesses -> Get document for: " + aPersonBusinessProcessID)
	findOptions := options.Find()
	findOptions.SetLimit(2)

	// Here's an array in which you can store the decoded documents
	var results []*PersonBusinessProcess

	collection := m.personBusinessProcessCollection
	// Passing bson.D{{}} as the filter matches all documents in the collection
	filter := bson.M{"referenceNumber": bson.D{{Key: "$eq", Value: aPersonBusinessProcessID}}}
	//filter := bson.D{{}}

	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		slog.Error(fmt.Sprintf("PersonBusinessProcessGet -> Filter Error: %s", err))
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem PersonBusinessProcess
		err := cur.Decode(&elem)
		if err != nil {
			slog.Error(fmt.Sprintf("PersonBusinessProcess Get -> Decode Error ->%s", err))
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		slog.Error(fmt.Sprintf("PersonBusinessProcessDB Get -> Cur Error ->%s", err))
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	if len(results) == 0 {
		return nil, errors.New("personBusinessProcess not found")
	}
	personBusinessProcess := results[0]
	slog.Debug("PersonBusinessProcesses -> Got document for: " + aPersonBusinessProcessID)
	return personBusinessProcess, err
}
func (m BusinessProcessMongoDB) InsertPersonBusinessProcess(personBusinessProcess *PersonBusinessProcess) error {
	slog.Debug("PersonBusinessProcessDB -> Inserting document for: " + personBusinessProcess.ReferenceNumber)

	collection := m.personBusinessProcessCollection

	insertResult, err := collection.InsertOne(context.TODO(), personBusinessProcess)
	if err != nil {
		slog.Error(fmt.Sprintf("PersonBusinessProcessDB -> Error: %s", err))
	}

	slog.Debug(fmt.Sprintln("Inserted a single document: ", insertResult.InsertedID))
	return nil
}
func (m BusinessProcessMongoDB) DeleteAllPersonBusinessProcesses() error {
	slog.Debug("PersonBusinessProcess -> Deleting All documents")

	deleteResult, err := m.personBusinessProcessCollection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		slog.Error(fmt.Sprintf("PersonBusinessProcessDB -> Error: %s", err))
	}
	s := fmt.Sprintf("Deleted %v documents in the PersonBusinessProcess collection\n", deleteResult.DeletedCount)
	slog.Debug(s)
	return nil
}

// UpdatePersonBusinessProcess inserts documents into the PersonBusinessProcess Collection
func (m *BusinessProcessMongoDB) UpdatePersonBusinessProcess(key string, personBusinessProcess *PersonBusinessProcess) error {
	slog.Debug("PersonBusinessProcessDB -> Updating document for: " + key)

	filter := bson.M{"referenceNumber": bson.D{{Key: "$eq", Value: key}}}

	updateResult, err := m.personBusinessProcessCollection.ReplaceOne(context.TODO(), filter, personBusinessProcess)
	if err != nil {
		slog.Error(fmt.Sprintf("PersonBusinessProcessDB -> Error: %s", err))
		return err
	}
	if updateResult.ModifiedCount == 0 {
		slog.Error(fmt.Sprintf("PersonBusinessProcessDB -> Error: %s", err))
		return errors.New("person business process not updated")
	}

	s := fmt.Sprintf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	slog.Debug(s)
	return nil
}
