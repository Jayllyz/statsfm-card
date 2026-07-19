package card

import (
	"fmt"
	"html"
	"strings"
)

// EscapeText mirrors the PHP validate_and_escape('string', ...) SVG text
// path: replace '&' with a fullwidth lookalike (raw '&' breaks SVG XML),
// then HTML-escape the rest.
func EscapeText(s string) string {
	s = strings.ReplaceAll(s, "&", "＆")
	return html.EscapeString(s)
}

// Rect returns a <rect>, optionally filled with a top-to-bottom linear
// gradient when both gradientStart and gradientStop are non-empty.
func Rect(x, y, width, height int, radius int, gradientStart, gradientStop string) string {
	if gradientStart == "" || gradientStop == "" {
		return fmt.Sprintf(
			`<rect x="%d" y="%d" width="%d" height="%d" rx="%d" ry="%d" />`,
			x, y, width, height, radius, radius,
		)
	}

	id := "grad-" + gradientStart + "-" + gradientStop
	defs := fmt.Sprintf(
		`<defs><linearGradient id="%s" x1="0%%" y1="0%%" x2="0%%" y2="100%%"><stop offset="0%%" style="stop-color:#%s;stop-opacity:1" /><stop offset="100%%" style="stop-color:#%s;stop-opacity:1" /></linearGradient></defs>`,
		id, gradientStart, gradientStop,
	)
	rect := fmt.Sprintf(
		`<rect x="%d" y="%d" width="%d" height="%d" fill="url(#%s)" rx="%d" ry="%d" />`,
		x, y, width, height, id, radius, radius,
	)

	return defs + rect
}

// Img returns an <image>, optionally clipped to a rounded-rect mask.
func Img(base64PNG string, x, y, width, height, radius int) string {
	image := fmt.Sprintf(
		`<image x="%d" y="%d" width="%d" height="%d" href="data:image/png;base64,%s" />`,
		x, y, width, height, base64PNG,
	)
	if radius == 0 {
		return image
	}

	id := fmt.Sprintf("mask-%d-%d", x, y)
	mask := fmt.Sprintf(
		`<defs><mask id="%s"><rect x="%d" y="%d" width="%d" height="%d" fill="white" rx="%d" ry="%d" /></mask></defs>`,
		id, x, y, width, height, radius, radius,
	)
	image = fmt.Sprintf(
		`<image x="%d" y="%d" width="%d" height="%d" href="data:image/png;base64,%s" mask="url(#%s)" />`,
		x, y, width, height, base64PNG, id,
	)

	return mask + image
}

// Text returns a <text> element. text is caller-escaped (see EscapeText).
func Text(text string, x, y int, color string, size int, weight, anchor string) string {
	return fmt.Sprintf(
		`<text x="%d" y="%d" fill="%s" style="text-anchor: %s; font-family: Arial; font-size: %dpx; font-weight: %s;">%s</text>`,
		x, y, color, anchor, size, weight, text,
	)
}

// Wrap wraps SVG content in the root <svg> element.
func Wrap(width, height int, content string) string {
	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="%d">%s</svg>`,
		width, height, content,
	)
}

// ErrorSVG renders a card showing a single centered error message.
func ErrorSVG(width, height, rounded int, gradientStart, gradientStop, message string) string {
	content := Rect(0, 0, width, height, rounded, gradientStart, gradientStop)
	content += Text(EscapeText(message), width/2, height/2, "white", 12, "bold", "middle")
	return Wrap(width, height, content)
}
