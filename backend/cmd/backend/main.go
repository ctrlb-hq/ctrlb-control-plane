package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/assets"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	database "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/db"
	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	frontendnode "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/node"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/joho/godotenv"
)

func main() {

	utils.InitLogger()

	_ = godotenv.Load()

	// Access your JWT secret from environment variables
	constants.JWT_SECRET = os.Getenv("JWT_SECRET")
	if constants.JWT_SECRET == "" {
		utils.Logger.Fatal("JWT_SECRET is not set in environment")
	}

	workerCountEnv := os.Getenv("WORKER_COUNT")
	if workerCountEnv != "" {
		count, err := strconv.Atoi(workerCountEnv)
		if err != nil {
			constants.WORKER_COUNT = 4
		} else {
			constants.WORKER_COUNT = count
		}
	} else {
		constants.WORKER_COUNT = 4
	}

	checkIntervalMinsEnv := os.Getenv("CHECK_INTERVAL_MINS")
	if checkIntervalMinsEnv != "" {
		count, err := strconv.Atoi(checkIntervalMinsEnv)
		if err != nil {
			constants.CHECK_INTERVAL_SEC = 10
		} else {
			constants.CHECK_INTERVAL_SEC = count
		}
	} else {
		constants.CHECK_INTERVAL_SEC = 10
	}

	if portEnv := os.Getenv("PORT"); portEnv != "" {
		constants.PORT = portEnv
	} else {
		constants.PORT = "8096" // Default value
	}

	// Read ENV from environment variable or set default
	if envEnv := os.Getenv("ENV"); envEnv != "" {
		constants.ENV = envEnv
	} else {
		constants.ENV = "prod" // Default value
	}

	db, err := database.DBInit("./backend.db")
	if err != nil {
		utils.Logger.Sugar().Fatal("Failed to initialize DB: %s", err)
		return
	}

	schemasFS, err := fs.Sub(assets.Schemas, "schemas")
	if err != nil {
		utils.Logger.Sugar().Errorf("Failed to initialize schema: %s", err)
	}

	uiSchemasFS, err := fs.Sub(assets.Schemas, "ui_schemas")
	if err != nil {
		utils.Logger.Sugar().Errorf("Failed to initialize UI schema: %s", err)
	}

	err = database.LoadSchemasFromDirectory(db, schemasFS, uiSchemasFS, database.GetComponentTypeMap(), database.GetSignalSupportMap())
	if err != nil {
		utils.Logger.Sugar().Fatalf("Failed to load component schemas: %v", err)
	}

	utils.Logger.Info("Component schemas loaded into database")

	agentQueueRepository := queue.NewQueueRepository(db)

	agentQueue := queue.NewQueue(constants.WORKER_COUNT, constants.CHECK_INTERVAL_SEC, agentQueueRepository)

	if err = agentQueue.RefreshMonitoring(); err != nil {
		utils.Logger.Fatal("Unable to update existing agent")
		return
	}

	agentRepository := agent.NewAgentRepository(db)
	authRepository := auth.NewAuthRepository(db)

	frontendAgentRepository := frontendagent.NewFrontendAgentRepository(db)
	frontendPipelineRepository := frontendpipeline.NewFrontendPipelineRepository(db)
	frontendNodeRepository := frontendnode.NewFrontendNodeRepository(db)

	frontendAgentService := frontendagent.NewFrontendAgentService(frontendAgentRepository, agentQueue)
	frontendPipelineService := frontendpipeline.NewFrontendPipelineService(frontendPipelineRepository)
	frontendNodeService := frontendnode.NewFrontendNodeService(frontendNodeRepository)

	agentService := agent.NewAgentService(agentRepository, agentQueue, frontendPipelineService)
	authService := auth.NewAuthService(authRepository)

	handler := api.NewHandler(agentService, authService, frontendAgentService, frontendPipelineService, frontendNodeService)

	router := api.NewRouter(handler)

	handlerWithCors := middleware.CorsMiddleware(router)

	server := &http.Server{
		Addr:    ":" + constants.PORT,
		Handler: handlerWithCors,
	}

	go func() {
		utils.Logger.Info(fmt.Sprintf("Server started on: %s", constants.PORT))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			utils.Logger.Sugar().Fatal("Failed to start Server: %s", err)
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the server
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	<-interruptChan
	utils.Logger.Info("Received interrupt signal, shutting down...")
}
