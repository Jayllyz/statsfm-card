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
	path := fmt.Sprintf("users/%s/top/%s", url.PathEscape(username), url.PathEscape(itemType))
	query := url.Values{"range": {rangeParam}, "limit": {strconv.Itoa(limit)}}

	var top TopResponse
	if err := c.getJSON(ctx, path, query, &top); err != nil {
		return nil, err
	}

	return &top, nil
}

// StreamStats fetches a user's aggregate listening totals (total played
// time and stream count) for the given range.
func (c *Client) StreamStats(ctx context.Context, username, rangeParam string) (*StreamStats, error) {
	path := fmt.Sprintf("users/%s/streams/stats", url.PathEscape(username))
	query := url.Values{"range": {rangeParam}}

	var stats StreamStats
	if err := c.getJSON(ctx, path, query, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// getJSON issues a GET request to baseURL/path with query, decoding a
// successful JSON response into out.
func (c *Client) getJSON(ctx context.Context, path string, query url.Values, out any) error {
	reqURL := fmt.Sprintf("%s/%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("statsfm: build request: %w", err)
	}

	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("statsfm: request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn().Err(closeErr).Msg("statsfm: close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("statsfm: unexpected status %d for %s", resp.StatusCode, reqURL)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("statsfm: decode response: %w", err)
	}

	return nil
}
