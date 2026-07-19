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

// ImageFetcher fetches and encodes the image at url, returning a base64 PNG.
type ImageFetcher func(ctx context.Context, url string) (base64PNG string, err error)

// Render composes items into the full SVG card described by params.
// Items beyond params.Limit are ignored. Images that fail to fetch are
// logged and skipped, leaving the name/stat text in place.
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

		body.WriteString(Text(EscapeText(name), centerX, artistTextY, "white", 9, "normal", "middle"))

		if stat := statText(item, params.Display); stat != "" {
			body.WriteString(Text(EscapeText(stat), centerX, statTextY, "white", 9, "bold", "middle"))
		}
	}

	content := Rect(0, 0, params.Width, params.Height, params.Rounded, params.GStart, params.GStop) + body.String()

	return Wrap(params.Width, params.Height, content)
}

// statText formats the played-time or stream-count label for an item,
// matching the PHP hours/streams display modes (rounded hours, space as
// the thousands separator).
func statText(item statsfm.Item, display string) string {
	switch {
	case display == config.DisplayHours && item.PlayedMs != nil:
		hours := int64(math.Round(float64(*item.PlayedMs) / 1000 / 60 / 60))
		return formatThousands(hours) + " h"
	case display == config.DisplayStreams && item.Streams != nil:
		return formatThousands(*item.Streams) + " s"
	default:
		return ""
	}
}

// formatThousands renders n with a space as the thousands separator,
// matching PHP's number_format($n, 0, '.', ' ').
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
