package main

import (
	"log"
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
	frontendconfig "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/config"
	frontendconfigV2 "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/configV2"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/queue"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Access your JWT secret from environment variables
	constants.JWT_SECRET = os.Getenv("JWT_SECRET")
	if constants.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET is not set in .env file")
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

	db, err := database.InitializeDB()
	if err != nil {
		return
	}

	agentQueue := queue.NewQueue(constants.WORKER_COUNT, db)
	agentQueue.StartStatusCheck()

	agentRepository := agent.NewAgentRepository(db)
	authRepository := auth.NewAuthRepository(db)
	frontendAgentRepository := frontendagent.NewFrontendAgentRepository(db)
	frontendPipelineRepository := frontendpipeline.NewFrontendPipelineRepository(db)
	frontendConfigRepository := frontendconfig.NewFrontendConfigRepositoryV2(db)
	frontendConfigRepositoryV2 := frontendconfigV2.NewFrontendConfigRepository(db)

	agentService := agent.NewAgentService(agentRepository, agentQueue)
	authService := auth.NewAuthService(authRepository)
	frontendAgentService := frontendagent.NewFrontendAgentService(frontendAgentRepository, agentQueue)
	frontendPipelineService := frontendpipeline.NewFrontendPipelineService(frontendPipelineRepository, agentQueue)
	frontendConfigService := frontendconfig.NewFrontendAgentService(frontendConfigRepository)
	frontendConfigServiceV2 := frontendconfigV2.NewFrontendAgentService(frontendConfigRepositoryV2)

	router := api.NewRouter(agentService, authService, frontendAgentService, frontendPipelineService, frontendConfigService, frontendConfigServiceV2)

	handlerWithCors := middleware.CorsMiddleware(router)

	server := &http.Server{
		Addr:    ":" + constants.PORT,
		Handler: handlerWithCors,
	}

	go func() {
		log.Println("Server started on :", constants.PORT)
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
