package shutdown

import (
	"context"
	"net/http"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

var Server *http.Server

func ShutdownServer() {
	// Check if server was initialized before attempting shutdown
	if Server == nil {
		logger.Logger.Info("HTTP server not initialized, skipping shutdown")
		return
	}
	logger.Logger.Info("Shuting Down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Server.Shutdown(shutdownCtx); err != nil {
		logger.Logger.Sugar().Errorf("Error occured while shutting down HTTP server: %v", err)
	}
}
