package cache

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLRUSetGet(t *testing.T) {
	t.Parallel()

	c := New(10, time.Hour)
	c.Set("a", []byte("svg-a"))

	got, ok := c.Get("a")
	if !ok || string(got) != "svg-a" {
		t.Errorf("Get(%q) = (%q, %v), want (%q, true)", "a", got, ok, "svg-a")
	}
}

func TestLRUGetMiss(t *testing.T) {
	t.Parallel()

	c := New(10, time.Hour)

	_, ok := c.Get("missing")
	if ok {
		t.Error("Get() on missing key = true, want false")
	}
}

func TestLRUOverwrite(t *testing.T) {
	t.Parallel()

	c := New(10, time.Hour)
	c.Set("a", []byte("v1"))
	c.Set("a", []byte("v2"))

	got, ok := c.Get("a")
	if !ok || string(got) != "v2" {
		t.Errorf("Get(%q) = (%q, %v), want (%q, true)", "a", got, ok, "v2")
	}
	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
}

func TestLRUExpiry(t *testing.T) {
	t.Parallel()

	c := New(10, time.Minute)

	fakeNow := time.Now()
	c.now = func() time.Time { return fakeNow }

	c.Set("a", []byte("svg-a"))

	fakeNow = fakeNow.Add(2 * time.Minute)

	_, ok := c.Get("a")
	if ok {
		t.Error("Get() on expired key = true, want false")
	}
	if c.Len() != 0 {
		t.Errorf("Len() after expiry = %d, want 0 (Get evicts expired entries)", c.Len())
	}
}

func TestLRUEvictsLeastRecentlyUsed(t *testing.T) {
	t.Parallel()

	c := New(2, time.Hour)
	c.Set("a", []byte("va"))
	c.Set("b", []byte("vb"))

	// Touch "a" so "b" becomes the least recently used.
	c.Get("a")

	c.Set("c", []byte("vc"))

	if _, ok := c.Get("b"); ok {
		t.Error("Get(\"b\") = true, want false: least recently used entry should have been evicted")
	}
	if _, ok := c.Get("a"); !ok {
		t.Error("Get(\"a\") = false, want true: recently used entry should survive eviction")
	}
	if _, ok := c.Get("c"); !ok {
		t.Error("Get(\"c\") = false, want true: newly inserted entry should be present")
	}
	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2", c.Len())
	}
}

func TestLRUConcurrentAccess(t *testing.T) {
	t.Parallel()

	c := New(50, time.Hour)

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := strconv.Itoa(i % 20)
			c.Set(key, []byte(key))
			c.Get(key)
		}(i)
	}
	wg.Wait()

	if c.Len() > 50 {
		t.Errorf("Len() = %d, want <= capacity 50", c.Len())
	}
}
