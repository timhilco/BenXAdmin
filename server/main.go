package main

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"log/slog"
	"os"
	"server/application"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	opts := slog.HandlerOptions{
		//Level: slog.LevelInfo,
		Level: slog.LevelDebug,
	}
	ev := datatypes.EnvironmentVariables{
		TemplateDirectory: "./templates/",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)

	// CORS is enabled only in prod profile
	ev.Cors = os.Getenv("profile") == "prod"
	ev.IsKafka = true
	personMongoDB := db.NewPersonMongo()
	businessProcessMongoDB := businessProcess.NewBusinessProcessMongo()
	planMongoDB := db.NewPlanMongo()
	defer personMongoDB.CloseClientDB()
	defer businessProcessMongoDB.CloseClientDB()
	defer planMongoDB.CloseClientDB()
	application.StartConsumer(personMongoDB, businessProcessMongoDB, planMongoDB)
	appVersion := 2
	var app application.Application
	switch appVersion {
	case 1:
		app = application.NewApplication(personMongoDB, businessProcessMongoDB, ev)
	case 2:
		app = application.NewApplication2(personMongoDB, businessProcessMongoDB, planMongoDB, ev)
	}

	err := app.Serve()
	slog.Error("Error", err)

}

func clientOptions() *options.ClientOptions {
	host := "db"
	if os.Getenv("profile") != "prod" {
		host = "localhost"
	}
	return options.Client().ApplyURI(
		"mongodb://" + host + ":27017",
	)
}
