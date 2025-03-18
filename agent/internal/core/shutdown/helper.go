package shutdown

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

var Server *http.Server

func ShutdownServer(wg *sync.WaitGroup) {
	log.Println("Shuting Down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
	wg.Wait()
}
