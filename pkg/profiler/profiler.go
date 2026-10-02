package profiler

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof" // import pprof
	"os"
)

// Start runs the pprof HTTP server on the given port if ENV is not production.
func Start(port string) {
	if os.Getenv("ENV") != "production" {
		go func() {
			slog.Info("Starting pprof server", "port", port)
			if err := http.ListenAndServe(":"+port, nil); err != nil {
				slog.Error("pprof server failed", "error", err)
			}
		}()
	}
}
