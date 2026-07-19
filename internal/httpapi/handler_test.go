package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Jayllyz/statsfm-card/internal/cache"
	"github.com/Jayllyz/statsfm-card/internal/statsfm"
	"github.com/rs/zerolog"
)

func newTestHandler(t *testing.T, statsBody string, statsStatus int, imgHandler http.HandlerFunc) (*Handler, *httptest.Server, *httptest.Server) {
	t.Helper()

	statsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statsStatus)
		_, _ = w.Write([]byte(statsBody))
	}))
	t.Cleanup(statsSrv.Close)

	var imgSrv *httptest.Server
	if imgHandler != nil {
		imgSrv = httptest.NewServer(imgHandler)
		t.Cleanup(imgSrv.Close)
	}

	client := statsfm.New(statsSrv.URL, zerolog.Nop())
	h := New(client, http.DefaultClient, cache.New(10, time.Hour), zerolog.Nop())

	return h, statsSrv, imgSrv
}

func TestHandlerNoDataFound(t *testing.T) {
	t.Parallel()

	h, _, _ := newTestHandler(t, `{"items":[]}`, http.StatusOK, nil)

	req := httptest.NewRequest(http.MethodGet, "/?username=nobody", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (errors render as SVG, not HTTP errors)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "No data found") {
		t.Errorf("body = %q, want error message", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("Content-Type = %q, want image/svg+xml", ct)
	}
}

func TestHandlerAPIError(t *testing.T) {
	t.Parallel()

	h, _, _ := newTestHandler(t, `{}`, http.StatusInternalServerError, nil)

	req := httptest.NewRequest(http.MethodGet, "/?username=nobody", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Error fetching data") {
		t.Errorf("body = %q, want API error message", rec.Body.String())
	}
}

func TestHandlerSuccessAndCacheHit(t *testing.T) {
	t.Parallel()

	imgHandler := func(w http.ResponseWriter, r *http.Request) {
		// 1x1 PNG.
		png := []byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
			0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
			0x42, 0x60, 0x82,
		}
		_, _ = w.Write(png)
	}

	imgSrv := httptest.NewServer(http.HandlerFunc(imgHandler))
	t.Cleanup(imgSrv.Close)

	body := `{"items":[{"artist":{"name":"Muse","image":"` + imgSrv.URL + `/muse.png"}}]}`
	h, _, _ := newTestHandler(t, body, http.StatusOK, nil)

	req := httptest.NewRequest(http.MethodGet, "/?username=sheldon", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Muse") {
		t.Fatalf("body = %q, want item name", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "<image") {
		t.Errorf("body missing <image>: %q", rec.Body.String())
	}

	// Second request with the same URL should be served from cache.
	req2 := httptest.NewRequest(http.MethodGet, "/?username=sheldon", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	if rec2.Body.String() != rec.Body.String() {
		t.Errorf("cached response differs from original: %q vs %q", rec2.Body.String(), rec.Body.String())
	}
}
