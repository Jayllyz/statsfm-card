package statsfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/rs/zerolog"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return New(srv.URL, zerolog.Nop())
}

func TestClientTopItemsSuccess(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotQuery url.Values

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"artist":{"name":"Muse","image":"muse.png"}}]}`))
	})

	got, err := client.TopItems(context.Background(), "sheldon_cooper", "artists", "lifetime", 5)
	if err != nil {
		t.Fatalf("TopItems() error = %v", err)
	}

	wantPath := "/users/sheldon_cooper/top/artists"
	if gotPath != wantPath {
		t.Errorf("request path = %q, want %q", gotPath, wantPath)
	}
	if gotQuery.Get("range") != "lifetime" || gotQuery.Get("limit") != "5" {
		t.Errorf("request query = %v, want range=lifetime limit=5", gotQuery)
	}

	if len(got.Items) != 1 || got.Items[0].Artist == nil || got.Items[0].Artist.Name != "Muse" {
		t.Errorf("TopItems() = %+v, unexpected result", got)
	}
}

func TestClientTopItemsNonOKStatus(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.TopItems(context.Background(), "user", "artists", "lifetime", 5)
	if err == nil {
		t.Fatal("TopItems() error = nil, want error on non-200 status")
	}
}

func TestClientTopItemsInvalidJSON(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})

	_, err := client.TopItems(context.Background(), "user", "artists", "lifetime", 5)
	if err == nil {
		t.Fatal("TopItems() error = nil, want decode error")
	}
}

func TestClientTopItemsUsernameEscaped(t *testing.T) {
	t.Parallel()

	var gotPath string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	_, err := client.TopItems(context.Background(), "weird/user name", "artists", "lifetime", 5)
	if err != nil {
		t.Fatalf("TopItems() error = %v", err)
	}

	want := "/users/weird/user name/top/artists"
	if gotPath != want {
		t.Errorf("request path = %q, want %q (net/http unescapes PathEscape on the way in)", gotPath, want)
	}
}
