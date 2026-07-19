// Command server runs the statsfm-card HTTP service.
package main

import (
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
	cacheCapacity     = 500
	statsfmHTTPClient = 10 * time.Second
	imageHTTPClient   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	statsClient := statsfm.New(config.StatsfmBaseURL, logger)
	imageClient := &http.Client{Timeout: imageHTTPClient}
	c := cache.New(cacheCapacity, config.CacheTTL)

	handler := httpapi.New(statsClient, imageClient, c, logger)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.Info().Str("addr", addr).Msg("starting statsfm-card server")

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("server stopped")
	}
}
