package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api"
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

	if err := godotenv.Load(); err != nil {
		utils.Logger.Fatal("Error loading .env file")
	}

	// Access your JWT secret from environment variables
	constants.JWT_SECRET = os.Getenv("JWT_SECRET")
	if constants.JWT_SECRET == "" {
		utils.Logger.Fatal("JWT_SECRET is not set in .env file")
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

	db, err := database.DBInit()
	if err != nil {
		utils.Logger.Fatal(fmt.Sprintf("Failed to initialize DB: %s", err))
		return
	}

	agentQueue := queue.NewQueue(constants.WORKER_COUNT, db)
	agentQueue.StartStatusCheck()

	agentRepository := agent.NewAgentRepository(db)
	authRepository := auth.NewAuthRepository(db)

	frontendAgentRepositoryV2 := frontendagent.NewFrontendAgentRepository(db)
	frontendPipelineRepositoryV2 := frontendpipeline.NewFrontendPipelineRepository(db)
	frontendNodeRepositoryV2 := frontendnode.NewFrontendNodeRepository(db)

	agentService := agent.NewAgentService(agentRepository, agentQueue)
	authService := auth.NewAuthService(authRepository)

	frontendAgentServiceV2 := frontendagent.NewFrontendAgentService(frontendAgentRepositoryV2, agentQueue)
	frontendPipelineServiceV2 := frontendpipeline.NewFrontendPipelineService(frontendPipelineRepositoryV2, agentQueue)
	frontendNodeServiceV2 := frontendnode.NewFrontendNodeService(frontendNodeRepositoryV2)

	router := api.NewRouter(agentService, authService, frontendAgentServiceV2, frontendPipelineServiceV2, frontendNodeServiceV2)

	handlerWithCors := middleware.CorsMiddleware(router)

	server := &http.Server{
		Addr:    ":" + constants.PORT,
		Handler: handlerWithCors,
	}

	go func() {
		utils.Logger.Info(fmt.Sprintf("Server started on: %s", constants.PORT))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			utils.Logger.Fatal(fmt.Sprintf("Failed to start Server: %s", err))
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the server
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	<-interruptChan
	utils.Logger.Info("Received interrupt signal, shutting down...")
}
