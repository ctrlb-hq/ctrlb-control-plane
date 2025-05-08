package shutdown_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/shutdown"
)

func TestShutdownServer_Success(t *testing.T) {
	// Create a test server that blocks on request
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	})

	srv := &http.Server{
		Addr:    ":0", // let OS pick a free port
		Handler: mux,
	}

	// Assign to the global Server variable
	shutdown.Server = srv

	go func() {
		_ = srv.ListenAndServe()
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Call the function under test
	shutdown.ShutdownServer()

	// If we reach here without panic or hang, it's a pass
	// (you could add a mock logger if needed to assert log output)
}
