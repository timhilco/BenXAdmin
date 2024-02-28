package db

import (
	"benefitsDomain/domain/person"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *PersonMongoDB) GetPersons() ([]*person.Person, error) {
	res, err := m.personCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		slog.Error("Error while fetching persons:" + err.Error())
		return nil, err
	}
	var person []*person.Person
	err = res.All(context.TODO(), &person)
	if err != nil {
		slog.Error("Error while decoding persons:" + err.Error())
		return nil, err
	}
	return person, nil
}

// GetPerson gets documents from the Person Collection
func (m *PersonMongoDB) GetPerson(aPersonID string, idType string) (*person.Person, error) {
	// Pass these options to the Find method
	slog.Debug("Person -> Get document for: " + aPersonID)
	findOptions := options.Find()
	findOptions.SetLimit(2)

	// Here's an array in which you can store the decoded documents
	var results []*person.Person

	collection := m.personCollection
	// Passing bson.D{{}} as the filter matches all documents in the collection
	var filter primitive.M
	switch idType {
	case "External":
		filter = bson.M{"externalId": bson.D{{Key: "$eq", Value: aPersonID}}}
	default:
		filter = bson.M{"internalId": bson.D{{Key: "$eq", Value: aPersonID}}}
	}
	//filter := bson.D{{}}

	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		slog.Error(fmt.Sprintf("PersonGet -> Filter Error: %s", err))
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem person.Person
		err := cur.Decode(&elem)
		if err != nil {
			slog.Error(fmt.Sprintf("Person Get -> Decode Error ->%s", err))
		}
		elem.ConvertContactPointsToStructs()
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		slog.Error(fmt.Sprintf("PersonDB Get -> Cur Error ->%s", err))
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	if len(results) == 0 {
		return nil, errors.New("person not found")
	}
	person := results[0]
	return person, nil
}
func (m *PersonMongoDB) InsertPerson(person *person.Person) error {
	slog.Debug("PersonDB -> Inserting document for: " + person.InternalId)

	collection := m.personCollection

	insertResult, err := collection.InsertOne(context.TODO(), person)
	if err != nil {
		slog.Error(fmt.Sprintf("PersonDB -> Error: %s", err))
	}

	slog.Debug(fmt.Sprintln("Inserted a single document: ", insertResult.InsertedID))
	return nil
}
func (m *PersonMongoDB) DeleteAllPersons() error {
	slog.Debug("Person -> Deleting All documents")

	deleteResult, err := m.personCollection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		slog.Error(fmt.Sprintf("PersonDB -> Error: %s", err))
	}
	s := fmt.Sprintf("Deleted %v documents in the Person collection\n", deleteResult.DeletedCount)
	slog.Debug(s)
	return nil
}

/*
// UpdatePerson inserts documents into the Person Collection
func (m *PersonMongoDB) UpdatePerson(ctx context.Context, key string, person *Person) error {
	p.logger.Debug("PersonDB -> Updating document for: " + key)
	Person.WaitingExpectationsID = deconstructWaitingExpecations(Person)
	filter := bson.M{"internalID": bson.D{{Key: "$eq", Value: key}}}

	collection := p.MongoClient.Database("PersonDB").Collection("Person")
	updateResult, err := collection.ReplaceOne(context.TODO(), filter, Person)
	if err != nil {
		p.logger.Fatal(fmt.Sprintf("PersonDB -> Error: %s", err))
	}

	s := fmt.Sprintf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	p.logger.Debug(s)
	return nil
}

*/
