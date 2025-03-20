package shutdown

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg"
)

var Server *http.Server

func ShutdownServer(wg *sync.WaitGroup) {
	pkg.Logger.Info("Shuting Down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Server.Shutdown(shutdownCtx); err != nil {
		pkg.Logger.Error(fmt.Sprintf("HTTP server Shutdown: %v", err))
	}
	wg.Wait()
}
