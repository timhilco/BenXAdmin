package db

import (
	"benefitsDomain/domain/benefitPlan"
	"context"
	"errors"
	"fmt"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *PlanMongoDB) GetBenefits() ([]*benefitPlan.Benefit, error) {
	res, err := m.benefitCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		slog.Error("Error while fetching benefits:" + err.Error())
		return nil, err
	}
	var benefit []*benefitPlan.Benefit
	err = res.All(context.TODO(), &benefit)
	if err != nil {
		slog.Error("Error while decoding benefits:" + err.Error())
		return nil, err
	}
	return benefit, nil
}

// GetBenefit gets documents from the Benefit Collection
func (m *PlanMongoDB) GetBenefit(aBenefitID string) (*benefitPlan.Benefit, error) {
	// Pass these options to the Find method
	slog.Debug("Benefit -> Get document for: " + aBenefitID)
	findOptions := options.Find()
	findOptions.SetLimit(2)

	// Here's an array in which you can store the decoded documents
	var results []*benefitPlan.Benefit

	collection := m.benefitCollection
	// Passing bson.D{{}} as the filter matches all documents in the collection
	filter := bson.M{"benefitId": bson.D{{Key: "$eq", Value: aBenefitID}}}
	//filter := bson.D{{}}

	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		slog.Error(fmt.Sprintf("BenefitGet -> Filter Error: %s", err))
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem benefitPlan.Benefit
		err := cur.Decode(&elem)
		if err != nil {
			slog.Error(fmt.Sprintf("Benefit Get -> Decode Error ->%s", err))
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		slog.Error(fmt.Sprintf("BenefitDB Get -> Cur Error ->%s", err))
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	if len(results) == 0 {
		return nil, errors.New("benefit not found")
	}
	benefit := results[0]
	return benefit, nil
}
func (m *PlanMongoDB) InsertBenefit(benefit *benefitPlan.Benefit) error {
	slog.Debug("BenefitDB -> Inserting document for: " + benefit.InternalId)

	collection := m.benefitCollection

	insertResult, err := collection.InsertOne(context.TODO(), benefit)
	if err != nil {
		slog.Error(fmt.Sprintf("BenefitDB -> Error: %s", err))
	}

	slog.Debug(fmt.Sprintln("Inserted a single document: ", insertResult.InsertedID))
	return nil
}
func (m *PlanMongoDB) DeleteAllBenefits() error {
	slog.Debug("Benefit -> Deleting All documents")

	deleteResult, err := m.benefitCollection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		slog.Error(fmt.Sprintf("BenefitDB -> Error: %s", err))
	}
	s := fmt.Sprintf("Deleted %v documents in the Benefit collection\n", deleteResult.DeletedCount)
	slog.Debug(s)
	return nil
}

// UpdateBenefit inserts documents into the Benefit Collection
func (p *PlanMongoDB) UpdateBenefit(ctx context.Context, key string, benefit *benefitPlan.Benefit) error {
	slog.Debug("BenefitDB -> Updating document for: " + key)
	filter := bson.M{"internalID": bson.D{{Key: "$eq", Value: key}}}

	collection := p.benefitCollection
	updateResult, err := collection.ReplaceOne(context.TODO(), filter, benefit)
	if err != nil {
		slog.Error(fmt.Sprintf("BenefitDB -> Error: %s", err))
	}

	s := fmt.Sprintf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	slog.Debug(s)
	return nil
}
