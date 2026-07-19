// Command server runs the statsfm-card HTTP service.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Jayllyz/statsfm-card/internal/cache"
	"github.com/Jayllyz/statsfm-card/internal/config"
	"github.com/Jayllyz/statsfm-card/internal/httpapi"
	"github.com/Jayllyz/statsfm-card/internal/statsfm"
	"github.com/rs/zerolog"
)

const (
	cacheCapacity      = 500
	statsfmHTTPClient  = 10 * time.Second
	imageHTTPClient    = 10 * time.Second
	readHeaderTimeout  = 5 * time.Second
	healthCheckTimeout = 2 * time.Second
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthCheck())
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	statsClient := statsfm.New(config.StatsfmBaseURL, logger)
	imageClient := &http.Client{Timeout: imageHTTPClient}
	c := cache.New(cacheCapacity, config.CacheTTL)

	handler := httpapi.New(statsClient, imageClient, c, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/", handler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.Info().Str("addr", addr).Msg("starting statsfm-card server")

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("server stopped")
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// runHealthCheck lets the same static binary double as the Docker
// HEALTHCHECK probe, since the distroless image has no shell/curl/wget.
// It returns a process exit code rather than calling os.Exit directly
// so deferred cleanup (context cancel, body close) always runs.
func runHealthCheck() int {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	host := addr
	if addr[0] == ':' {
		host = "127.0.0.1" + addr
	}

	url := fmt.Sprintf("http://%s/health", host)

	ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) //nolint:gosec // fixed localhost target built from our own LISTEN_ADDR, not user input
	if err != nil {
		return 1
	}

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // see NewRequestWithContext comment above
	if err != nil {
		return 1
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close in a short-lived CLI probe

	if resp.StatusCode != http.StatusOK {
		return 1
	}

	return 0
}
