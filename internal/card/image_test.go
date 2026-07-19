package card

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
)

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}

	return buf.Bytes()
}

func solidImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}

	return img
}

func TestCropToSquare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		w, h     int
		wantSide int
	}{
		{"already square", 100, 100, 100},
		{"wider than tall", 200, 100, 100},
		{"taller than wide", 100, 200, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := solidImage(tt.w, tt.h, color.White)
			got := cropToSquare(src)

			b := got.Bounds()
			if b.Dx() != tt.wantSide || b.Dy() != tt.wantSide {
				t.Errorf("cropToSquare(%dx%d) bounds = %v, want %dx%d", tt.w, tt.h, b, tt.wantSide, tt.wantSide)
			}
		})
	}
}

func TestFetchSquarePNGSuccess(t *testing.T) {
	t.Parallel()

	src := solidImage(40, 20, color.RGBA{R: 255, A: 255})
	data := encodePNG(t, src)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)

	got, err := FetchSquarePNG(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("FetchSquarePNG() error = %v", err)
	}

	if got == "" {
		t.Fatal("FetchSquarePNG() returned empty string")
	}
}

func TestFetchSquarePNGNonOKStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	_, err := FetchSquarePNG(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("FetchSquarePNG() error = nil, want error on non-200 status")
	}
}

func TestFetchSquarePNGInvalidImage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not an image"))
	}))
	t.Cleanup(srv.Close)

	_, err := FetchSquarePNG(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("FetchSquarePNG() error = nil, want decode error")
	}
}
