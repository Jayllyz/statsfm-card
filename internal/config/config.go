// Package config holds request parameter defaults and cache settings
// for the statsfm-card service.
package config

import "time"

const (
	// CacheTTL is how long a rendered card stays fresh before regeneration.
	CacheTTL = 24 * time.Hour
	// NotFoundImage is served when stats.fm has no image for an item.
	NotFoundImage = "https://upload.wikimedia.org/wikipedia/commons/4/49/A_black_image.jpg"
	// StatsfmBaseURL is the stats.fm API root used to fetch top items.
	StatsfmBaseURL = "https://beta-api.stats.fm/api/v1"
)

// Params holds the card's query-string driven configuration, with
// defaults matching the original PHP DEFAULT_PARAMS.
type Params struct {
	Username string
	Range    string
	Type     string
	Display  string
	Limit    int
	Width    int
	Height   int
	Spacing  int
	YOffset  int
	Rounded  int
	IRounded int
	GStart   string
	GStop    string
}

// Default returns the baseline Params before query-string overrides are applied.
func Default() Params {
	return Params{
		Username: "",
		Range:    "lifetime",
		Type:     "artists",
		Display:  "hours",
		Limit:    5,
		Width:    580,
		Height:   180,
		Spacing:  20,
		YOffset:  12,
		Rounded:  10,
		IRounded: 4,
		GStart:   "0D1117",
		GStop:    "000000",
	}
}
