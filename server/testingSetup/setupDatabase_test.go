package setupDatabase

import (
	"benefitsDomain/domain/db"
	"benefitsDomain/random"
	"fmt"
	"log/slog"
	"os"
	"testing"
)

func Test_Setup_Population_Scenario_1(t *testing.T) {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	n := 10
	count := 5
	personMongoDB := db.NewPersonMongo()
	personMongoDB.DeleteAllPersons()
	personMongoDB.DeleteAllWorkers()
	personMongoDB.DeleteAllParticipants()
	defer personMongoDB.CloseClientDB()
	for i := n; i < n+count; i++ {
		options := make(map[string]interface{})
		eid := fmt.Sprintf("0%d-%d-00%d", i, i, i)
		options["ExternalId"] = eid
		person := random.CreateFakeItPerson(options)
		personMongoDB.InsertPerson(person)
		options["WorkerId"] = "W" + eid
		options["PersonInternalId"] = person.InternalId
		worker := random.CreateFakeItWorker(options)
		personMongoDB.InsertWorker(worker)
		slog.Info("Created Person: " + eid)

	}

}
