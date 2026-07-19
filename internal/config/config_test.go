package config

import "testing"

func TestDefault(t *testing.T) {
	t.Parallel()

	got := Default()
	want := Params{
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

	if got != want {
		t.Errorf("Default() = %+v, want %+v", got, want)
	}
}
