// Package card renders stats.fm top-items data into an SVG card.
package card

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif" // register decoders for formats stats.fm may serve
	_ "image/jpeg"
	"image/png"
	"net/http"
)

// FetchSquarePNG downloads the image at url, crops it to a centered
// square (matching Intervention\Image's cover()), and returns it as a
// base64-encoded PNG suitable for an SVG <image> href.
func FetchSquarePNG(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("card: build image request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("card: fetch image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("card: unexpected status %d fetching %s", resp.StatusCode, url)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("card: decode image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, cropToSquare(img)); err != nil {
		return "", fmt.Errorf("card: encode image: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// cropToSquare returns the centered, smaller-side square crop of img,
// matching Intervention\Image's cover() for a square target.
func cropToSquare(img image.Image) image.Image {
	b := img.Bounds()
	side := min(b.Dx(), b.Dy())

	offsetX := b.Min.X + (b.Dx()-side)/2
	offsetY := b.Min.Y + (b.Dy()-side)/2

	dst := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(dst, dst.Bounds(), img, image.Pt(offsetX, offsetY), draw.Src)

	return dst
}
