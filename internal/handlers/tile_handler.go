package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/kwaabs/m-events/internal/auth"
)

// TileHandler proxies MVT tile requests to Martin after JWT validation.
// Supports both Authorization header and ?token= query param since MapLibre
// cannot send custom headers on tile requests.
type TileHandler struct {
	martinURL string
	jwtSvc    *auth.JWTService
	valkey    *auth.ValkeyService
}

func NewTileHandler(jwtSvc *auth.JWTService, valkey *auth.ValkeyService) *TileHandler {
	url := os.Getenv("MARTIN_URL")
	if url == "" {
		url = "http://localhost:9401"
	}
	url = strings.TrimRight(url, "/")
	return &TileHandler{martinURL: url, jwtSvc: jwtSvc, valkey: valkey}
}

// Source names match what Martin auto-publishes from the view names.
var validSources = map[string]bool{
	"customer_map_view":      true,
	"CustomerRecords":        true, // ← add this line
	"district_event_summary": true,
}

// ProxyTile handles GET /api/v1/tiles/{source}/{z}/{x}/{y}
// Auth: Bearer token in Authorization header OR ?token= query param.
func (h *TileHandler) ProxyTile(w http.ResponseWriter, r *http.Request) {
	source := r.PathValue("source")
	z := r.PathValue("z")
	x := r.PathValue("x")
	y := r.PathValue("y")

	if source == "" || z == "" || x == "" || y == "" {
		writeError(w, http.StatusBadRequest, "invalid tile path", nil)
		return
	}

	if !validSources[source] {
		writeError(w, http.StatusNotFound, fmt.Sprintf("unknown tile source %q", source), nil)
		return
	}

	// ── Auth ─────────────────────────────────────────────────────────────────
	// The Authenticate middleware already validated the Bearer header if present.
	// For tile requests from MapLibre we also accept ?token= query param.
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		// Try ?token= query param (MapLibre tile URL pattern)
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			writeTileUnauthorized(w)
			return
		}
		parsed, err := h.jwtSvc.Validate(tokenStr)
		if err != nil {
			writeTileUnauthorized(w)
			return
		}
		// Check blacklist
		if parsed.ID != "" {
			revoked, _ := h.valkey.IsRevoked(r.Context(), parsed.ID)
			if revoked {
				writeTileUnauthorized(w)
				return
			}
		}
		claims = parsed
	}
	_ = claims // available for future per-user tile filtering

	// ── Proxy to Martin ───────────────────────────────────────────────────────
	targetURL := fmt.Sprintf("%s/%s/%s/%s/%s", h.martinURL, source, z, x, y)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build tile request", err)
		return
	}
	req.Header.Set("Accept", r.Header.Get("Accept"))
	if enc := r.Header.Get("Accept-Encoding"); enc != "" {
		req.Header.Set("Accept-Encoding", enc)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("martin unreachable", "url", targetURL, "error", err)
		writeError(w, http.StatusBadGateway, "tile server unreachable", err)
		return
	}
	defer resp.Body.Close()

	// Forward Martin response headers
	for key, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(key, v)
		}
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/vnd.mapbox-vector-tile")
	}
	w.Header().Set("Cache-Control", "public, max-age=60")

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// MartinHealth proxies Martin's health/catalog endpoint so you can verify
// tile sources without hitting Martin directly.
// GET /api/v1/tiles/health
func (h *TileHandler) MartinHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(h.martinURL + "/catalog")
	if err != nil {
		writeError(w, http.StatusBadGateway, "martin unreachable", err)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeTileUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
