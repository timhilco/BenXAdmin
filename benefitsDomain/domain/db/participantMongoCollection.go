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

func (m PersonMongoDB) GetParticipants(config map[string]string) ([]*personRoles.Participant, error) {
	res, err := m.participantCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		slog.Error("Error while fetching participants:" + err.Error())
		return nil, err
	}
	var participants []*personRoles.Participant
	var aPersonId, keyType string
	keyType = config["keyType"]
	aPersonId = config["PersonId"]
	var filter primitive.M
	var results []*personRoles.Participant
	switch keyType {

	case "PersonId":
		slog.Debug("Participant -> Get document for (External): " + aPersonId)
		filter = bson.M{"personId": bson.D{{Key: "$eq", Value: aPersonId}}}

		// Passing bson.D{{}} as the filter matches all documents in the participantCollection
		//filter := bson.D{{}}
		findOptions := options.Find()
		cur, err := m.participantCollection.Find(context.TODO(), filter, findOptions)
		if err != nil {
			slog.Error(fmt.Sprintf("ParticipantGet -> Filter Error: %s", err))
		}

		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		for cur.Next(context.TODO()) {

			// create a value into which the single document can be decoded
			var elem personRoles.Participant
			err := cur.Decode(&elem)
			if err != nil {
				slog.Error(fmt.Sprintf("Participant Get -> Decode Error ->%s", err))
			}

			results = append(results, &elem)
		}

		if err := cur.Err(); err != nil {
			slog.Error(fmt.Sprintf("ParticipantDB Get -> Cur Error ->%s", err))
		}

		// Close the cursor once finished
		cur.Close(context.TODO())
		if len(results) == 0 {
			return nil, errors.New("participant not found")
		}
		participants = results
	default:
		err = res.All(context.TODO(), &participants)
		if err != nil {
			slog.Error("Error while decoding participants:" + err.Error())
			return nil, err
		}
	}
	return participants, nil
}

// GetParticipant gets documents from the Participant Collection
func (m PersonMongoDB) GetParticipant(config map[string]string) (*personRoles.Participant, error) {
	// Pass these options to the Find method
	var aParticipantId, aPersonId, aBenefitId, keyType string
	keyType = config["keyType"]
	aParticipantId = config["ParticipantId"]
	aPersonId = config["PersonId"]
	aBenefitId = config["BenefitId"]

	slog.Debug("Participant -> Get document for: " + aParticipantId)
	findOptions := options.Find()
	findOptions.SetLimit(2)

	// Here's an array in which you can store the decoded documents
	var results []*personRoles.Participant

	participantCollection := m.participantCollection
	var filter primitive.M
	switch keyType {
	case "InternalId":
		slog.Debug("Participant -> Get document for (Internal): " + aParticipantId)
		filter = bson.M{"internalId": bson.D{{Key: "$eq", Value: aParticipantId}}}
	case "PersonId/BenefitId":
		slog.Debug("Participant -> Get document for (External): " + aPersonId + ":" + aBenefitId)
		filter = bson.M{"personId": bson.D{{Key: "$eq", Value: aPersonId}},
			"benefitId": bson.D{{Key: "$eq", Value: aBenefitId}},
		}
		// Passing bson.D{{}} as the filter matches all documents in the participantCollection
		//filter := bson.D{{}}
	}
	cur, err := participantCollection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		slog.Error(fmt.Sprintf("ParticipantGet -> Filter Error: %s", err))
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem personRoles.Participant
		err := cur.Decode(&elem)
		if err != nil {
			slog.Error(fmt.Sprintf("Participant Get -> Decode Error ->%s", err))
		}

		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		slog.Error(fmt.Sprintf("ParticipantDB Get -> Cur Error ->%s", err))
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	if len(results) == 0 {
		return nil, errors.New("participant not found")
	}
	participant := results[0]
	return participant, nil
}

/*
	func (m PersonMongoDB) GetParticipantByPersonBenefit(aPersonId string, aBenefitId string) (*personRoles.Participant, error) {
		// Pass these options to the Find method
		slog.Debug("Participant -> Get document for: " + aPersonId + "+" + aBenefitId)
		findOptions := options.Find()
		findOptions.SetLimit(2)

		// Here's an array in which you can store the decoded documents
		var results []*personRoles.Participant

		participantCollection := m.participantCollection
		// Passing bson.D{{}} as the filter matches all documents in the participantCollection
		filter := bson.M{"personId": bson.D{{Key: "$eq", Value: aPersonId}},
			"benefitId": bson.D{{Key: "$eq", Value: aBenefitId}},
		}
		//filter := bson.D{{}}

		cur, err := participantCollection.Find(context.TODO(), filter, findOptions)
		if err != nil {
			slog.Error(fmt.Sprintf("ParticipantGet -> Filter Error: %s", err))
		}

		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		for cur.Next(context.TODO()) {

			// create a value into which the single document can be decoded
			var elem personRoles.Participant
			err := cur.Decode(&elem)
			if err != nil {
				slog.Error(fmt.Sprintf("Participant Get -> Decode Error ->%s", err))
			}

			results = append(results, &elem)
		}

		if err := cur.Err(); err != nil {
			slog.Error(fmt.Sprintf("ParticipantDB Get -> Cur Error ->%s", err))
		}

		// Close the cursor once finished
		cur.Close(context.TODO())
		if len(results) == 0 {
			return nil, errors.New("participant not found")
		}
		participant := results[0]
		return participant, nil
	}
*/
func (m PersonMongoDB) InsertParticipant(participant *personRoles.Participant) error {
	slog.Debug("ParticipantDB -> Inserting document for: " + participant.InternalId)

	participantCollection := m.participantCollection

	insertResult, err := participantCollection.InsertOne(context.TODO(), participant)
	if err != nil {
		slog.Error(fmt.Sprintf("ParticipantDB -> Error: %s", err))
	}

	slog.Debug(fmt.Sprintln("Inserted a single document: ", insertResult.InsertedID))
	return nil
}
func (m PersonMongoDB) DeleteAllParticipants() error {
	slog.Debug("Participant -> Deleting All documents")

	deleteResult, err := m.participantCollection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		slog.Error(fmt.Sprintf("ParticipantDB -> Error: %s", err))
	}
	s := fmt.Sprintf("Deleted %v documents in the Participant participantCollection\n", deleteResult.DeletedCount)
	slog.Debug(s)
	return nil
}

// UpdateParticipant updates documents into the Participant Collection
func (m *PersonMongoDB) UpdateParticipant(key string, participant *personRoles.Participant) error {
	slog.Debug("ParticipantDB -> Updating document for: " + key)
	filter := bson.M{"internalId": bson.D{{Key: "$eq", Value: key}}}

	participantCollection := m.participantCollection
	updateResult, err := participantCollection.ReplaceOne(context.TODO(), filter, participant)
	if err != nil {
		slog.Error(fmt.Sprintf("ParticipantDB -> Error: %s", err))
	}

	s := fmt.Sprintf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	slog.Debug(s)
	return nil
}
