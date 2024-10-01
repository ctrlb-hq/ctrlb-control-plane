package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ctrlb-hq/all-father/internal/api"
	"github.com/ctrlb-hq/all-father/internal/repositories"
	"github.com/ctrlb-hq/all-father/internal/services"
	dbcreator "github.com/ctrlb-hq/all-father/pkg/db-creator"
	"github.com/joho/godotenv"
)

func main() {

	configFile := ".env"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	err := godotenv.Load(configFile)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	workerCount, _ := strconv.Atoi(os.Getenv("WORKER_COUNT"))

	db, err := dbcreator.DBCreator()
	if err != nil {
		return
	}

	agentQueue := services.NewQueue(workerCount)

	agentRepository := repositories.NewAgentRepository(db)

	agentService := services.NewAgentService(agentRepository, agentQueue)

	services := services.Services{
		AgentService: agentService,
	}

	handler := api.NewRouter(&services)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Println("Server started on :", port)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start Server:", err)
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the server
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	<-interruptChan
}
