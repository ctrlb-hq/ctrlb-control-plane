package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	sessionManager "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth/session-manager"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	database "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/db"
	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	frontendconfig "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/config"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/queue"
)

func main() {

	constants.WORKER_COUNT = *flag.Int("wc", 4, "Number of worker threads")
	constants.PORT = *flag.String("port", "8096", "Server port for communication")
	constants.ENV = *flag.String("env", "prod", "For testing purpose")

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
	frontendConfigRepository := frontendconfig.NewFrontendConfigRepository(db)

	authSessionManager := sessionManager.NewSessionManager()
	agentService := agent.NewAgentService(agentRepository, agentQueue)
	authService := auth.NewAuthService(authRepository, authSessionManager)
	frontendAgentService := frontendagent.NewFrontendAgentService(frontendAgentRepository, agentQueue)
	frontendPipelineService := frontendpipeline.NewFrontendPipelineService(frontendPipelineRepository, agentQueue)
	frontendConfigService := frontendconfig.NewFrontendAgentService(frontendConfigRepository)

	router := api.NewRouter(agentService, authService, frontendAgentService, frontendPipelineService, frontendConfigService, authService.SessionManager)

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
