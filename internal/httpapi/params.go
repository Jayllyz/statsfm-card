// Package httpapi wires the HTTP entrypoint: parses request params,
// checks the cache, fetches from stats.fm, and renders the SVG card.
package httpapi

import (
	"net/url"
	"strconv"

	"github.com/Jayllyz/statsfm-card/internal/card"
	"github.com/Jayllyz/statsfm-card/internal/config"
)

// parseParams overlays query values onto the default Params, matching
// the PHP behavior of only overriding keys present in the request and
// falling back to the default on invalid integers.
func parseParams(q url.Values) config.Params {
	p := config.Default()

	if v := q.Get("username"); v != "" {
		p.Username = v
	}

	if v := q.Get("range"); v != "" {
		p.Range = v
	}

	if v := q.Get("type"); v != "" {
		p.Type = v
	}

	if v := q.Get("display"); v != "" {
		p.Display = v
	}

	if v := q.Get("g_start"); v != "" {
		p.GStart = card.EscapeText(v)
	}

	if v := q.Get("g_stop"); v != "" {
		p.GStop = card.EscapeText(v)
	}

	p.Limit = intParam(q, "limit", p.Limit)
	p.Width = intParam(q, "width", p.Width)
	p.Height = intParam(q, "height", p.Height)
	p.Spacing = intParam(q, "spacing", p.Spacing)
	p.YOffset = intParam(q, "y_offset", p.YOffset)
	p.Rounded = intParam(q, "rounded", p.Rounded)
	p.IRounded = intParam(q, "i_rounded", p.IRounded)

	return p
}

// intParam parses key from q, falling back to def if absent or not a
// valid integer (matching PHP's validate_and_escape('int', ...)).
func intParam(q url.Values, key string, def int) int {
	v, ok := q[key]
	if !ok || len(v) == 0 {
		return def
	}

	n, err := strconv.Atoi(v[0])
	if err != nil {
		return def
	}

	return n
}
