package card

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Jayllyz/statsfm-card/internal/config"
	"github.com/Jayllyz/statsfm-card/internal/statsfm"
	"github.com/rs/zerolog"
)

func fakeFetcher(fail bool) ImageFetcher {
	return func(_ context.Context, url string) (string, error) {
		if fail {
			return "", errors.New("boom")
		}
		return "base64-for-" + url, nil
	}
}

func TestFormatThousands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{5, "5"},
		{999, "999"},
		{1000, "1 000"},
		{1234567, "1 234 567"},
		{-1234, "-1 234"},
	}

	for _, tt := range tests {
		if got := formatThousands(tt.in); got != tt.want {
			t.Errorf("formatThousands(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStatText(t *testing.T) {
	t.Parallel()

	playedMs := int64(3_600_000 * 2) // 2 hours exactly
	streams := int64(4200)

	tests := []struct {
		name    string
		item    statsfm.Item
		display string
		want    string
	}{
		{"hours", statsfm.Item{PlayedMs: &playedMs}, "hours", "2 h"},
		{"streams", statsfm.Item{Streams: &streams}, "streams", "4 200 s"},
		{"hours display but no playedMs", statsfm.Item{}, "hours", ""},
		{"unknown display", statsfm.Item{PlayedMs: &playedMs}, "minutes", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := statText(tt.item, tt.display); got != tt.want {
				t.Errorf("statText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderBuildsExpectedItems(t *testing.T) {
	t.Parallel()

	params := config.Default()
	params.Type = "artists"
	params.Limit = 2

	playedMs := int64(3_600_000)
	items := []statsfm.Item{
		{Artist: &statsfm.Entity{Name: "Muse", Image: "muse.png"}, PlayedMs: &playedMs},
		{Artist: &statsfm.Entity{Name: "Radiohead", Image: "radiohead.png"}, PlayedMs: &playedMs},
		{Artist: &statsfm.Entity{Name: "Should Be Ignored", Image: "x.png"}, PlayedMs: &playedMs},
	}

	got := Render(context.Background(), params, items, fakeFetcher(false), zerolog.Nop())

	if !strings.HasPrefix(got, "<svg") || !strings.HasSuffix(got, "</svg>") {
		t.Fatalf("Render() not a well-formed svg: %q", got)
	}
	if !strings.Contains(got, "Muse") || !strings.Contains(got, "Radiohead") {
		t.Errorf("Render() missing expected item names: %q", got)
	}
	if strings.Contains(got, "Should Be Ignored") {
		t.Errorf("Render() included item beyond limit: %q", got)
	}
	if !strings.Contains(got, "base64-for-muse.png") {
		t.Errorf("Render() missing fetched image data: %q", got)
	}
	if !strings.Contains(got, "1 h") {
		t.Errorf("Render() missing formatted stat text: %q", got)
	}
}

func TestRenderSkipsImageOnFetchError(t *testing.T) {
	t.Parallel()

	params := config.Default()
	params.Limit = 1
	items := []statsfm.Item{
		{Artist: &statsfm.Entity{Name: "Muse", Image: "muse.png"}},
	}

	got := Render(context.Background(), params, items, fakeFetcher(true), zerolog.Nop())

	if strings.Contains(got, "<image") {
		t.Errorf("Render() should skip <image> on fetch error: %q", got)
	}
	if !strings.Contains(got, "Muse") {
		t.Errorf("Render() should still render item name: %q", got)
	}
}

func TestRenderFallsBackToNotFoundImage(t *testing.T) {
	t.Parallel()

	params := config.Default()
	params.Limit = 1
	items := []statsfm.Item{
		{Artist: &statsfm.Entity{Name: "No Image"}},
	}

	var gotURL string
	fetch := func(_ context.Context, url string) (string, error) {
		gotURL = url
		return "b64", nil
	}

	Render(context.Background(), params, items, fetch, zerolog.Nop())

	if gotURL != config.NotFoundImage {
		t.Errorf("Render() fetched %q, want fallback %q", gotURL, config.NotFoundImage)
	}
}
