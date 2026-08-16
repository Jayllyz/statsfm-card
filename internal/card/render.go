package card

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/Jayllyz/statsfm-card/internal/config"
	"github.com/Jayllyz/statsfm-card/internal/statsfm"
	"github.com/rs/zerolog"
)

// imageSize is the fixed square size, in pixels, of each item's cover art.
const imageSize = 80

// labelColor and labelSize style every text label on the card (item names,
// stats, and the total-time footer).
const (
	labelColor = "white"
	labelSize  = 9
)

// ImageFetcher fetches and encodes the image at url, returning a base64 PNG.
type ImageFetcher func(ctx context.Context, url string) (base64PNG string, err error)

// label returns a card text element styled with labelColor/labelSize,
// center-anchored at x,y with the given font weight.
func label(text string, x, y int, weight string) string {
	return Text(EscapeText(text), x, y, labelColor, labelSize, weight, "middle")
}

// Render composes items into the full SVG card described by params.
// Items beyond params.Limit are ignored. Images that fail to fetch are
// logged and skipped, leaving the name/stat text in place. When
// params.TotalMs is set, a total-time footer is added at the bottom.
func Render(ctx context.Context, params config.Params, items []statsfm.Item, fetch ImageFetcher, logger zerolog.Logger) string {
	limit := min(params.Limit, len(items))

	startX := (params.Width - (imageSize*limit + params.Spacing*(limit-1))) / 2
	startY := (params.Height - imageSize) / 2

	var body strings.Builder

	for i, item := range items[:limit] {
		localYOffset := (i % 2) * params.YOffset
		localStartY := startY - localYOffset
		localStartX := startX + (imageSize+params.Spacing)*i
		artistTextY := localStartY - 5
		statTextY := localStartY + imageSize + 12
		centerX := localStartX + imageSize/2

		name, imageURL := item.NameAndImage(params.Type)
		if imageURL == "" {
			imageURL = config.NotFoundImage
		}

		base64PNG, err := fetch(ctx, imageURL)
		if err != nil {
			logger.Warn().Err(err).Str("url", imageURL).Msg("card: fetch item image failed, skipping")
		} else {
			body.WriteString(Img(base64PNG, localStartX, localStartY, imageSize, imageSize, params.IRounded))
		}

		body.WriteString(label(name, centerX, artistTextY, "normal"))

		if stat := statText(item, params.Display); stat != "" {
			body.WriteString(label(stat, centerX, statTextY, "bold"))
		}
	}

	content := Rect(0, 0, params.Width, params.Height, params.Rounded, params.GStart, params.GStop) + body.String()

	if params.TotalMs != nil {
		content += label("Total "+formatDuration(*params.TotalMs), params.Width/2, params.Height-8, "bold")
	}

	return Wrap(params.Width, params.Height, content)
}

// formatDuration renders a millisecond duration as a whole-hours label
// (space thousands separator), matching the "h"-suffixed style used for
// per-item stats.
func formatDuration(ms int64) string {
	hours := int64(math.Round(float64(ms) / 1000 / 60 / 60))
	return formatThousands(hours) + " h"
}

// statText formats the played-time or stream-count label for an item
// (rounded hours, space as the thousands separator).
func statText(item statsfm.Item, display string) string {
	switch {
	case display == config.DisplayHours && item.PlayedMs != nil:
		return formatDuration(*item.PlayedMs)
	case display == config.DisplayStreams && item.Streams != nil:
		return formatThousands(*item.Streams) + " s"
	default:
		return ""
	}
}

// formatThousands renders n with a space as the thousands separator.
func formatThousands(n int64) string {
	s := strconv.FormatInt(n, 10)

	neg := ""
	if s[0] == '-' {
		neg, s = "-", s[1:]
	}

	var out []byte

	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}

		out = append(out, c)
	}

	return neg + string(out)
}
