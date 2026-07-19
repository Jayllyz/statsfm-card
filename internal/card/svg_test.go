package card

import (
	"strings"
	"testing"
)

func TestEscapeText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, in, want string
	}{
		{"plain", "Muse", "Muse"},
		{"ampersand", "Simon & Garfunkel", "Simon ＆ Garfunkel"},
		{"html special chars", `<script>"'`, "&lt;script&gt;&#34;&#39;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := EscapeText(tt.in); got != tt.want {
				t.Errorf("EscapeText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRectNoGradient(t *testing.T) {
	t.Parallel()

	got := Rect(0, 0, 580, 180, 10, "", "")
	want := `<rect x="0" y="0" width="580" height="180" rx="10" ry="10" />`
	if got != want {
		t.Errorf("Rect() = %q, want %q", got, want)
	}
}

func TestRectWithGradient(t *testing.T) {
	t.Parallel()

	got := Rect(0, 0, 580, 180, 10, "0D1117", "000000")
	if !strings.Contains(got, "<linearGradient") {
		t.Errorf("Rect() with gradient colors missing <linearGradient>: %q", got)
	}
	if !strings.Contains(got, "stop-color:#0D1117") || !strings.Contains(got, "stop-color:#000000") {
		t.Errorf("Rect() missing gradient stop colors: %q", got)
	}
	if !strings.Contains(got, `fill="url(#`) {
		t.Errorf("Rect() missing gradient fill reference: %q", got)
	}
}

func TestImgNoRadius(t *testing.T) {
	t.Parallel()

	got := Img("QUJD", 10, 20, 80, 80, 0)
	want := `<image x="10" y="20" width="80" height="80" href="data:image/png;base64,QUJD" />`
	if got != want {
		t.Errorf("Img() = %q, want %q", got, want)
	}
}

func TestImgWithRadius(t *testing.T) {
	t.Parallel()

	got := Img("QUJD", 10, 20, 80, 80, 4)
	if !strings.Contains(got, "<mask") {
		t.Errorf("Img() with radius missing <mask>: %q", got)
	}
	if !strings.Contains(got, `mask="url(#`) {
		t.Errorf("Img() missing mask reference on <image>: %q", got)
	}
}

func TestText(t *testing.T) {
	t.Parallel()

	got := Text("Muse", 50, 60, "white", 9, "normal", "middle")
	want := `<text x="50" y="60" fill="white" style="text-anchor: middle; font-family: Arial; font-size: 9px; font-weight: normal;">Muse</text>`
	if got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	got := Wrap(580, 180, "<rect />")
	want := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="580" height="180"><rect /></svg>`
	if got != want {
		t.Errorf("Wrap() = %q, want %q", got, want)
	}
}

func TestErrorSVG(t *testing.T) {
	t.Parallel()

	got := ErrorSVG(580, 180, 10, "0D1117", "000000", "[500] Error fetching data")
	if !strings.HasPrefix(got, "<svg") || !strings.HasSuffix(got, "</svg>") {
		t.Errorf("ErrorSVG() not a well-formed svg wrapper: %q", got)
	}
	if !strings.Contains(got, "[500] Error fetching data") {
		t.Errorf("ErrorSVG() missing error message: %q", got)
	}
	if !strings.Contains(got, "<linearGradient") {
		t.Errorf("ErrorSVG() missing gradient background: %q", got)
	}
}
