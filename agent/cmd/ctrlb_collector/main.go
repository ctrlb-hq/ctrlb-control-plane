package main

import (
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/api"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/client"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/config"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/shutdown"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/filewatcher"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/systeminfo"
	"github.com/joho/godotenv"
)

func main() {
	logger.InitLogger()

	var wg sync.WaitGroup
	_ = godotenv.Load()

	// AGENT_CONFIG_PATH: Optional, with default fallback
	constants.AGENT_CONFIG_PATH = os.Getenv("AGENT_CONFIG_PATH")
	if constants.AGENT_CONFIG_PATH == "" {
		constants.AGENT_CONFIG_PATH = "./config.yaml"
	}

	// BACKEND_URL: Mandatory, no default
	constants.BACKEND_URL = os.Getenv("BACKEND_URL")
	if constants.BACKEND_URL == "" {
		logger.Logger.Fatal("BACKEND_URL environment variable is not set. Exiting...")
	}

	constants.PIPELINE_NAME = os.Getenv("PIPELINE_NAME")
	if constants.PIPELINE_NAME == "" {
		logger.Logger.Info("PIPELINE_NAME environment variable is not set. Using default value: empty string.")
	}

	constants.STARTED_BY = os.Getenv("STARTED_BY")
	if constants.STARTED_BY == "" {
		logger.Logger.Info("STARTED_BY environment variable is not set. Using default value: empty string.")
	}

	constants.PORT = os.Getenv("PORT")
	if constants.PORT == "" {
		logger.Logger.Info("PORT environment variable is not set. Using default value: 3421.")
		constants.PORT = "3421"
	}

	constants.AGENT_TYPE = os.Getenv("AGENT_TYPE")
	if constants.AGENT_TYPE == "" {
		logger.Logger.Info("AGENT_TYPE environment variable is not set. Using default value: otel.")
		constants.AGENT_TYPE = "otel"
	}

	isValid := slices.Contains(constants.SUPPORTED_AGENT_TYPES, constants.AGENT_TYPE)
	if !isValid {
		logger.Logger.Sugar().Infof("Invalid AGENT_TYPE: %s. Using default value: otel.", constants.AGENT_TYPE)
		constants.AGENT_TYPE = "otel"
	}

	// Configure heartbeat interval
	if heartbeatIntervalEnv := os.Getenv("HEARTBEAT_INTERVAL_SEC"); heartbeatIntervalEnv != "" {
		if interval, err := strconv.Atoi(heartbeatIntervalEnv); err == nil && interval > 0 {
			constants.HEARTBEAT_INTERVAL_SEC = interval
		}
	}

	// Check if config file exists
	if _, err := os.Stat(constants.AGENT_CONFIG_PATH); err != nil {
		logger.Logger.Sugar().Errorf("Config file doesn't exist at location: %v", constants.AGENT_CONFIG_PATH)
		logger.Logger.Fatal("Exiting....")
	}

	adapter, err := adapters.NewAdapter(&wg, constants.AGENT_TYPE)
	if err != nil {
		logger.Logger.Sugar().Fatalf("Failed to create adapter: %v", err)
	}

	// 3. Start the agent
	err = adapter.Initialize()
	if err != nil {
		logger.Logger.Sugar().Fatalf("Failed to start Agent adapter: %v", err)
	}
	logger.Logger.Info("Agent started successfully")

	go filewatcher.WatchFile(constants.AGENT_CONFIG_PATH, adapter)

	version, err := adapter.GetVersion()
	if err != nil {
		logger.Logger.Sugar().Errorf("Error while fetching agent version: %v", err)
		adapter.GracefulShutdown()
		os.Exit(1)
	}
	constants.AGENT_VERSION = version

	// Channel to signal fatal errors from goroutines
	fatalErrChan := make(chan error, 2)

	// Call Backend server which will be informed about agent being started
	wg.Add(1)
	go func() {
		defer wg.Done()
		sys := systeminfo.NewSystemInfo()
		httpClient := &http.Client{
			Timeout: 20 * time.Second,
		}
		serverStartConfig, err := client.InformBackendServerStart(sys, httpClient)
		if err != nil {
			logger.Logger.Sugar().Errorf("Failed to register with backend server: %v", err)
			fatalErrChan <- err
			return
		}
		err = config.SaveToYAML(serverStartConfig, constants.AGENT_CONFIG_PATH)
		if err != nil {
			logger.Logger.Sugar().Errorf("Error writing config to file: %v", err)
			fatalErrChan <- err
			return
		}
		logger.Logger.Info("Successfully registered with the backend server")
	}()

	// Start heartbeat manager to periodically send status/metrics to backend
	heartbeatManager := client.NewHeartbeatManager(constants.HEARTBEAT_INTERVAL_SEC)
	heartbeatManager.Start()

	operatorService := operators.NewOperatorService(adapter, constants.AGENT_TYPE)

	handler := api.NewRouter(operatorService)

	server := &http.Server{
		Addr:    ":" + constants.PORT,
		Handler: handler,
	}

	//Used for shutting down server
	shutdown.Server = server

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Logger.Sugar().Infof("Client started at port: %s", constants.PORT)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Logger.Sugar().Errorf("Failed to start Server: %v", err)
			fatalErrChan <- err
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal or fatal error
	select {
	case <-sigChan:
		logger.Logger.Info("Received termination signal. Initiating graceful shutdown...")
	case err := <-fatalErrChan:
		logger.Logger.Sugar().Errorf("Fatal error occurred: %v. Initiating graceful shutdown...", err)
	}

	heartbeatManager.Stop()
	shutdown.ShutdownServer()
	adapter.GracefulShutdown()
	wg.Wait()
}
