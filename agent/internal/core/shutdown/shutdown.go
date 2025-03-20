package shutdown

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

var Server *http.Server

func ShutdownServer(wg *sync.WaitGroup) {
	logger.Logger.Info("Shuting Down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Server.Shutdown(shutdownCtx); err != nil {
		logger.Logger.Error(fmt.Sprintf("HTTP server Shutdown: %v", err))
	}
	wg.Wait()
}
