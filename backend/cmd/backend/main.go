package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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

	// Configure agent inactive timeout
	inactiveTimeoutEnv := os.Getenv("AGENT_INACTIVE_TIMEOUT_SEC")
	if inactiveTimeoutEnv != "" {
		count, err := strconv.Atoi(inactiveTimeoutEnv)
		if err == nil {
			constants.AGENT_INACTIVE_TIMEOUT_SEC = count
		}
	}

	// Configure staleness check interval
	stalenessCheckEnv := os.Getenv("STALENESS_CHECK_SEC")
	if stalenessCheckEnv != "" {
		count, err := strconv.Atoi(stalenessCheckEnv)
		if err == nil {
			constants.STALENESS_CHECK_SEC = count
		}
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

	uiSchemasFS, err := fs.Sub(assets.UI_Schemas, "ui_schemas")
	if err != nil {
		utils.Logger.Sugar().Errorf("Failed to initialize UI schema: %s", err)
	}

	err = database.LoadSchemasFromDirectory(db, schemasFS, uiSchemasFS, database.GetComponentTypeMap(), database.GetSignalSupportMap())
	if err != nil {
		utils.Logger.Sugar().Fatalf("Failed to load component schemas: %v", err)
	}

	utils.Logger.Info("Component schemas loaded into database")

	metricsRepository := queue.NewQueueRepository(db)

	// Start staleness checker for marking inactive agents
	stalenessChecker := queue.NewStalenessChecker(
		metricsRepository,
		constants.STALENESS_CHECK_SEC,
		constants.AGENT_INACTIVE_TIMEOUT_SEC,
	)
	stalenessChecker.Start()

	agentRepository := agent.NewAgentRepository(db)
	authRepository := auth.NewAuthRepository(db)

	frontendAgentRepository := frontendagent.NewFrontendAgentRepository(db)
	frontendPipelineRepository := frontendpipeline.NewFrontendPipelineRepository(db)
	frontendNodeRepository := frontendnode.NewFrontendNodeRepository(db)

	frontendAgentService := frontendagent.NewFrontendAgentService(frontendAgentRepository)
	frontendPipelineService := frontendpipeline.NewFrontendPipelineService(frontendPipelineRepository, frontendAgentService)
	frontendNodeService := frontendnode.NewFrontendNodeService(frontendNodeRepository)

	agentService := agent.NewAgentService(agentRepository, metricsRepository, frontendPipelineService)
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

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	utils.Logger.Info("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		utils.Logger.Sugar().Errorf("HTTP server shutdown error: %v", err)
	} else {
		utils.Logger.Info("HTTP server shutdown completed")
	}

	// Shutdown staleness checker
	stalenessChecker.Stop()

	// Close database connection
	if err := db.Close(); err != nil {
		utils.Logger.Sugar().Errorf("Database close error: %v", err)
	} else {
		utils.Logger.Info("Database connection closed")
	}

	utils.Logger.Info("Shutdown complete, exiting")
}
