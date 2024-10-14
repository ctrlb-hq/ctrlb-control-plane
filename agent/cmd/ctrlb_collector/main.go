package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters/fluentbit"
	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters/otel"
	"github.com/ctrlb-hq/ctrlb-collector/internal/api"
	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/helper"
	"github.com/ctrlb-hq/ctrlb-collector/internal/services"
	"github.com/ctrlb-hq/ctrlb-collector/pkg/serverclient"
)

func main() {
	var wg sync.WaitGroup

	constants.AGENT_CONFIG_PATH = *flag.String("config", "./internal/resources/config/otel.yaml", "Path to the agent configuration file")
	constants.AGENT_TYPE = *flag.String("type", "otel", "Type of the agent")
	constants.BACKEND_URL = *flag.String("backend", "http://pipeline.ctrlb.ai/", "URL of the backend server")
	constants.PORT = *flag.String("port", "443", "Agent port for communication with server")
	constants.ENV = *flag.String("env", "prod", "For testing purpose")
	flag.Parse()

	if _, err := os.Stat(constants.AGENT_CONFIG_PATH); err != nil {
		log.Fatal("Config file doesn't exist. Exiting....")
	}

	var adapter adapters.Adapter

	switch constants.AGENT_TYPE {
	case "fluent-bit":
		adapter = fluentbit.NewFluentBitAdapter(&wg)
	case "otel":
		adapter = otel.NewOTELCollectorAdapter(&wg)
	default:
		log.Fatal("Agent currently not supported. Exiting....")
		return
	}

	// Call Backend server which will be informed about agent being started
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := serverclient.InformBackendServerStart()
		if err != nil {
			log.Fatalf("failed to register with backend server: %v", err)
		} else {
			log.Println("successfully registered with the backend server")
		}
	}()

	// 3. Start the agent
	err := adapter.Initialize()
	if err != nil {
		log.Fatalf("Failed to start Agent adapter: %v", err)
	}

	log.Printf("%s agent started successfully", constants.AGENT_TYPE)

	operator_service := *services.NewOperatorService(adapter)
	if err != nil {
		log.Fatalf("Failed to initiate agent operator: %v", err)
	}

	var handler http.Handler

	handler = api.NewRouter(&operator_service)

	server := &http.Server{
		Addr:    ":" + constants.PORT,
		Handler: handler,
	}

	//Used for shutting down server
	helper.Server = server

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Client started at port:%s", constants.PORT)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start Server:", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan

	log.Printf("Received termination signal. Initiating graceful shutdown...")

	adapter.GracefulShutdown()

}
