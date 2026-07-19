# Plan: Rewrite statsfm-card in Go

## Current app (PHP)
- Single-file `index.php` HTTP handler: reads query params → calls stats.fm API (guzzle) → builds SVG string (rects/images/text) → embeds fetched artist/track images as base64 PNG (intervention/image, cropped square) → disk-caches SVG by md5(url) for 1 day
- `clean.php`: cron job, deletes cache files older than 7 days
- `constants.php`: default params + cache config
- Deployed via Docker (php8.4-apache + cron)

## Target Go architecture

```
statsfm-card/
├── cmd/server/main.go        # entrypoint, http server + graceful shutdown
├── internal/
│   ├── config/config.go      # default params, env-configurable cache dir/ttl
│   ├── httpapi/
│   │   ├── handler.go        # GET / handler, param parsing+validation
│   │   └── handler_test.go
│   ├── statsfm/
│   │   ├── client.go         # stats.fm API client (net/http, context, timeout)
│   │   └── types.go          # response structs (top items, artist/track/album)
│   ├── card/
│   │   ├── svg.go            # svg builder (rect, text, image nodes) via text/template or strings.Builder
│   │   ├── image.go          # fetch+decode image, square-crop via image/draw, encode png, base64
│   │   └── svg_test.go
│   └── cache/
│       ├── cache.go          # interface: Get/Set(key, svg []byte)
│       ├── lru.go            # in-memory LRU (size-bounded) + TTL check on Get
│       └── lru_test.go
├── go.mod
├── Dockerfile                # single static binary, distroless/scratch base
├── docker-compose.yml
└── README.md
```

## Logging
- `zerolog`, single global logger in `main.go`, injected into handler/cache via constructor
- Structured fields: `username`, `type`, `range`, `cache_hit`, `duration_ms`, `status`
- Replace PHP's silent `echo 'Image Error: ...'` with `log.Error().Err(err).Str("url", url).Msg("image fetch failed")`, still return error SVG to client

## Component mapping (PHP → Go)

| PHP | Go |
|---|---|
| guzzle client | `net/http.Client` w/ custom `Transport` (UA header, timeout); default TLS verify on (PHP's `verify=false` dropped — reverse proxy handles TLS, no reason to skip outbound verify) |
| intervention/image crop+encode | `image`, `image/draw` (crop to square), `image/png` (encode) — stdlib only, no dep needed |
| `addRect/addImg/addText` string concat | small SVG builder funcs, same signatures, return `template.HTML`-safe strings |
| `htmlspecialchars` escaping | `html.EscapeString` (stdlib) |
| query param parsing + defaults | struct w/ `url.Values` binding, explicit int/string validators mirroring `validate_and_escape` |
| md5(full URL) cache key | same: `md5.Sum([]byte(r.URL.String()))` |
| file-based cache w/ mtime check | in-memory LRU (`container/list` + map, or `hashicorp/golang-lru/v2`), entry = svg bytes + storedAt; TTL check (1 day) on Get, evict on Set when over capacity |
| `clean.php` cron + MAX_CACHE_TIME | not needed: LRU eviction + TTL-on-read handles staleness/size, no separate sweep process |
| apache + php-fpm container | plain Go binary, `net/http` server, no apache/cron needed |
| — | `zerolog` structured logging throughout (new, PHP had none) |

## Behavior to preserve exactly
- Same query params: `username, range, type, display, limit, width, height, spacing, y_offset, rounded, i_rounded, g_start, g_stop`
- Same defaults (constants.php values)
- Same SVG output shape (so existing embeds in READMEs don't break)
- Same cache semantics: 1 day TTL serve, 7 day max age purge
- Same error SVG on API failure / empty data
- `&` → `＆` workaround for SVG text safety

## Steps
1. `go mod init`, scaffold dirs above, add zerolog dep (+ lru lib if not hand-rolled)
2. Port `constants.php` → `internal/config` (struct + defaults)
3. Port stats.fm client + response types (`internal/statsfm`), w/ zerolog for request logging
4. Port image fetch/crop/encode (`internal/card/image.go`) using stdlib `image/*` — no external dep needed, drop intervention/image
5. Port SVG builders (`internal/card/svg.go`)
6. Build LRU cache (`internal/cache/lru.go`): capacity-bounded, TTL-on-read, thread-safe (mutex)
7. Wire HTTP handler (`internal/httpapi`), mount at `/`, zerolog middleware for access logs
8. Test coverage (target: high on cache + svg + param validation, the logic-heavy parts):
   - `cache`: set/get, TTL expiry, LRU eviction order, concurrent access (`-race`)
   - `card/svg`: exact string output for known inputs, escaping (`&` workaround), gradient/no-gradient paths
   - `card/image`: crop-to-square math, error path (bad image data)
   - `httpapi`: param parsing/defaults/validation, error-svg on API failure/empty data
   - `statsfm`: client against `httptest.Server` mocks
9. Dockerfile: multi-stage build → scratch/distroless, single binary, no apache/php/cron
10. Update docker-compose.yml, README (drop composer install instructions, add `docker run`/binary usage, drop cron-job cache-clean mention)
11. Manual verification: run locally, hit `?username=<test>`, diff SVG output against PHP version for same params
12. Remove PHP files once Go version verified in prod (index.php, clean.php, constants.php, composer.json, .htaccess, docker/apache.conf)

## Open questions
- LRU capacity: fixed entry count (e.g. 500) or byte-size bound? Need a number to size it right.
- Any consumers relying on PHP-specific error text/format that must match byte-for-byte?
- Deployment target unchanged (same Docker host), or opportunity to move to smaller distroless/scratch image?
