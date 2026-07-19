package httpapi

import (
	"net/url"
	"testing"

	"github.com/Jayllyz/statsfm-card/internal/config"
)

func TestParseParamsDefaults(t *testing.T) {
	t.Parallel()

	got := parseParams(url.Values{})
	want := config.Default()

	if got != want {
		t.Errorf("parseParams(empty) = %+v, want defaults %+v", got, want)
	}
}

func TestParseParamsOverrides(t *testing.T) {
	t.Parallel()

	q := url.Values{
		"username": {"sheldon_cooper"},
		"type":     {"tracks"},
		"range":    {"weeks"},
		"display":  {"streams"},
		"limit":    {"3"},
		"width":    {"800"},
		"g_start":  {"FF0000"},
	}

	got := parseParams(q)

	if got.Username != "sheldon_cooper" || got.Type != "tracks" || got.Range != "weeks" || got.Display != "streams" {
		t.Errorf("parseParams() string overrides not applied: %+v", got)
	}

	if got.Limit != 3 || got.Width != 800 {
		t.Errorf("parseParams() int overrides not applied: %+v", got)
	}

	if got.GStart != "FF0000" {
		t.Errorf("parseParams() GStart = %q, want %q", got.GStart, "FF0000")
	}
	// Untouched fields keep their default.
	if got.Height != config.Default().Height {
		t.Errorf("parseParams() Height = %d, want default %d", got.Height, config.Default().Height)
	}
}

func TestParseParamsInvalidIntFallsBackToDefault(t *testing.T) {
	t.Parallel()

	q := url.Values{"limit": {"not-a-number"}}

	got := parseParams(q)

	if got.Limit != config.Default().Limit {
		t.Errorf("parseParams() Limit = %d, want default %d on invalid input", got.Limit, config.Default().Limit)
	}
}

func TestParseParamsGStartEscaped(t *testing.T) {
	t.Parallel()

	q := url.Values{"g_start": {"<b>"}}

	got := parseParams(q)

	if got.GStart == "<b>" {
		t.Errorf("parseParams() GStart not escaped: %q", got.GStart)
	}
}
