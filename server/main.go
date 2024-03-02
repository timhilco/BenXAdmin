package main

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/businessProcess"
	"benefitsDomain/domain/db"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"server/application"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var SignalCh = make(chan os.Signal, 1)

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
	defer app.Close()
	router := app.GetRouter()
	ctx, cancel := context.WithCancel(context.Background())

	// Create a WaitGroup to keep track of running goroutines
	var wg sync.WaitGroup

	// Start the HTTP server
	wg.Add(1)
	go startHTTPServer(ctx, &wg, router)

	// Listen for termination signals

	signal.Notify(SignalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-SignalCh

	// Start the graceful shutdown process
	slog.Info("Gracefully shutting down HTTP server...")

	// Cancel the context to signal the HTTP server to stop
	cancel()

	// Wait for the HTTP server to finish
	wg.Wait()

	slog.Info("Shutdown complete.")

}
func startHTTPServer(ctx context.Context, wg *sync.WaitGroup, router *mux.Router) {
	defer wg.Done()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start the HTTP server in a separate goroutine
	go func() {
		slog.Info("Starting HTTP server on Port 3000..")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %s\n", err)
		}
	}()

	// Wait for the context to be canceled
	select {
	case <-ctx.Done():
		// Shutdown the server gracefully
		slog.Info("<- ctx.Done in startHTTPServer-- Shutting down HTTP server gracefully...")
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			slog.Info("HTTP server shutdown error: %s\n", err)
		}
	}

	slog.Info("HTTP server stopped.")
}

/*
	func clientOptions() *options.ClientOptions {
		host := "db"
		if os.Getenv("profile") != "prod" {
			host = "localhost"
		}
		return options.Client().ApplyURI(
			"mongodb://" + host + ":27017",
		)
	}
*/
func HandleShutdown(w http.ResponseWriter, r *http.Request) {
	SignalCh <- syscall.SIGTERM
}
