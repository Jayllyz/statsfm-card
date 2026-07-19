// Package statsfm provides a client for the stats.fm top-items API.
package statsfm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

// Client fetches top-items data from the stats.fm API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     zerolog.Logger
}

// New returns a Client targeting baseURL (e.g. config.StatsfmBaseURL).
func New(baseURL string, logger zerolog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		logger:     logger,
	}
}

// TopItems fetches a user's top items of the given type (artists, tracks,
// albums) for the given range, capped at limit results.
func (c *Client) TopItems(ctx context.Context, username, itemType, rangeParam string, limit int) (*TopResponse, error) {
	reqURL := fmt.Sprintf("%s/users/%s/top/%s", c.baseURL, url.PathEscape(username), url.PathEscape(itemType))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("statsfm: build request: %w", err)
	}

	q := req.URL.Query()
	q.Set("range", rangeParam)
	q.Set("limit", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("statsfm: request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn().Err(closeErr).Msg("statsfm: close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statsfm: unexpected status %d for %s", resp.StatusCode, reqURL)
	}

	var top TopResponse
	if err := json.NewDecoder(resp.Body).Decode(&top); err != nil {
		return nil, fmt.Errorf("statsfm: decode response: %w", err)
	}

	return &top, nil
}
