package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters/otel"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/agentcomm"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/api"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/config"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/operators"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/shutdownhelper"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
)

func main() {
	var wg sync.WaitGroup

	constants.AGENT_CONFIG_PATH = *flag.String("config", "./config.yaml", "Path to the agent configuration file")
	constants.AGENT_TYPE = *flag.String("type", "otel", "Type of the agent")
	constants.IS_PIPELINE = *flag.Bool("isPipeline", false, "Agent or Pipeline")
	constants.BACKEND_URL = *flag.String("backend", "http://pipeline.ctrlb.ai:8096", "URL of the backend server")
	constants.PORT = *flag.String("port", "443", "Agent port for communication with server")

	flag.Parse()

	if _, err := os.Stat(constants.AGENT_CONFIG_PATH); err != nil {
		log.Fatal("Config file doesn't exist. Exiting....")
	}

	var adapter adapters.Adapter
	adapter = otel.NewOTELCollectorAdapter(&wg)

	// 3. Start the agent
	err := adapter.Initialize()
	if err != nil {
		log.Fatalf("Failed to start Agent adapter: %v", err)
	}
	log.Printf("%s agent started successfully", constants.AGENT_TYPE)
	go config.WatchFile(constants.AGENT_CONFIG_PATH, adapter)

	version, err := adapter.GetVersion()
	if err != nil {
		log.Fatalln("error while fetching agent version: ", err)
	} else {
		constants.AGENT_VERSION = version
	}

	// Call Backend server which will be informed about agent being started
	wg.Add(1)
	go func() {
		defer wg.Done()
		agentWithConfig, err := agentcomm.InformBackendServerStart()
		if err != nil {
			log.Fatalf("failed to register with backend server: %v", err)
		} else {
			constants.AGENT = agentWithConfig
			configData := constants.AGENT.Config.Config
			err = utils.WriteConfigToFile(configData, constants.AGENT_CONFIG_PATH)
			if err != nil {
				log.Fatalf("error writing config to file: %v", err)
			}
			log.Println("successfully registered with the backend server")
		}
	}()

	operator_service := *operators.NewOperatorService(adapter)
	if &operator_service == nil {
		log.Fatalf("Failed to initiate agent operator")
	}

	var handler http.Handler

	handler = api.NewRouter(&operator_service)

	server := &http.Server{
		Addr:    ":" + constants.PORT,
		Handler: handler,
	}

	//Used for shutting down server
	shutdownhelper.Server = server

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Client started at port: %s", constants.PORT)
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
