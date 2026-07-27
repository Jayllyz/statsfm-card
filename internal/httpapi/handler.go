package httpapi

import (
	"context"
	"crypto/md5" //nolint:gosec // used only as a non-adversarial cache key, not for security
	"encoding/hex"
	"net/http"
	"time"

	"github.com/Jayllyz/statsfm-card/internal/cache"
	"github.com/Jayllyz/statsfm-card/internal/card"
	"github.com/Jayllyz/statsfm-card/internal/statsfm"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// Handler serves SVG top-items cards over HTTP.
type Handler struct {
	stats      *statsfm.Client
	httpClient *http.Client
	cache      *cache.LRU
	logger     zerolog.Logger
}

// New returns a Handler wired to the given collaborators.
func New(stats *statsfm.Client, httpClient *http.Client, c *cache.LRU, logger zerolog.Logger) *Handler {
	return &Handler{stats: stats, httpClient: httpClient, cache: c, logger: logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	w.Header().Set("Content-Type", "image/svg+xml")

	params := parseParams(r.URL.Query())
	cacheKey := requestCacheKey(r)

	logger := h.logger.With().
		Str("username", params.Username).
		Str("type", params.Type).
		Str("range", params.Range).
		Logger()

	if svg, ok := h.cache.Get(cacheKey); ok {
		writeSVG(w, svg)
		logger.Info().Bool("cache_hit", true).Dur("duration", time.Since(start)).Msg("served card")

		return
	}

	var top *statsfm.TopResponse

	g, gctx := errgroup.WithContext(r.Context())

	g.Go(func() error {
		var err error
		top, err = h.stats.TopItems(gctx, params.Username, params.Type, params.Range, params.Limit)
		return err
	})

	if params.Total {
		g.Go(func() error {
			stats, err := h.stats.StreamStats(gctx, params.Username, params.Range)
			if err != nil {
				logger.Warn().Err(err).Msg("stats.fm stream stats request failed, omitting total")
				return nil
			}

			params.TotalMs = &stats.Items.DurationMs

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.Error().Err(err).Msg("stats.fm request failed")
		writeSVG(w, []byte(card.ErrorSVG(params.Width, params.Height, params.Rounded, params.GStart, params.GStop, "[500] Error fetching data from API")))

		return
	}

	if len(top.Items) == 0 {
		writeSVG(w, []byte(card.ErrorSVG(params.Width, params.Height, params.Rounded, params.GStart, params.GStop, "[204] No data found")))
		return
	}

	svg := card.Render(r.Context(), params, top.Items, h.fetchImage, h.logger)
	h.cache.Set(cacheKey, []byte(svg))

	writeSVG(w, []byte(svg))
	logger.Info().Bool("cache_hit", false).Dur("duration", time.Since(start)).Msg("served card")
}

func (h *Handler) fetchImage(ctx context.Context, url string) (string, error) {
	return card.FetchSquarePNG(ctx, h.httpClient, url)
}

// requestCacheKey hashes the request path and query so that distinct
// parameter combinations get distinct cache entries.
func requestCacheKey(r *http.Request) string {
	sum := md5.Sum([]byte(r.URL.RequestURI())) //nolint:gosec // see import comment
	return hex.EncodeToString(sum[:])
}

func writeSVG(w http.ResponseWriter, svg []byte) {
	_, _ = w.Write(svg) //nolint:gosec // generated SVG bytes, not user-controlled HTML; gosec's taint check is a false positive here
}
