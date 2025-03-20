package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/api"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/client"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/config"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/shutdown"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
)

func main() {
	var wg sync.WaitGroup

	constants.AGENT_CONFIG_PATH = *flag.String("config", "./config.yaml", "Path to the agent configuration file")
	constants.BACKEND_URL = *flag.String("backend", "http://pipeline.ctrlb.ai:8096", "URL of the backend server")
	constants.PORT = *flag.String("port", "443", "Agent port for communication with server")

	flag.Parse()

	if _, err := os.Stat(constants.AGENT_CONFIG_PATH); err != nil {
		pkg.Logger.Fatal("Config file doesn't exist. Exiting....")
	}

	adapter, err := adapters.NewAdapter(&wg, constants.AGENT_TYPE)
	if err != nil {
		pkg.Logger.Fatal(fmt.Sprintf("Failed to create adapter: %v", err))
	}

	// 3. Start the agent
	err = adapter.Initialize()
	if err != nil {
		pkg.Logger.Fatal(fmt.Sprintf("Failed to start Agent adapter: %v", err))
	}
	pkg.Logger.Info(fmt.Sprintf("%s agent started successfully", constants.AGENT_TYPE))

	go config.WatchFile(constants.AGENT_CONFIG_PATH, adapter)

	version, err := adapter.GetVersion()
	if err != nil {
		pkg.Logger.Fatal(fmt.Sprintf("Error while fetching agent version: %v", err))
	} else {
		constants.AGENT_VERSION = version
	}

	// Call Backend server which will be informed about agent being started
	wg.Add(1)
	go func() {
		defer wg.Done()
		agentWithConfig, err := client.InformBackendServerStart()
		if err != nil {
			pkg.Logger.Fatal(fmt.Sprintf("Failed to register with backend server: %v", err))
		} else {
			constants.AGENT = agentWithConfig
			configData := constants.AGENT.Config.Config
			err = utils.WriteConfigToFile(configData, constants.AGENT_CONFIG_PATH)
			if err != nil {
				pkg.Logger.Fatal(fmt.Sprintf("Error writing config to file: %v", err))
			}
			pkg.Logger.Info("Successfully registered with the backend server")
		}
	}()

	operator_service := *operators.NewOperatorService(adapter)

	handler := api.NewRouter(&operator_service)

	server := &http.Server{
		Addr:    ":" + constants.PORT,
		Handler: handler,
	}

	//Used for shutting down server
	shutdown.Server = server

	wg.Add(1)
	go func() {
		defer wg.Done()
		pkg.Logger.Info(fmt.Sprintf("Client started at port: %s", constants.PORT))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			pkg.Logger.Fatal(fmt.Sprintf("Failed to start Server: %v", err))
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan

	pkg.Logger.Info("Received termination signal. Initiating graceful shutdown...")

	adapter.GracefulShutdown()

}
