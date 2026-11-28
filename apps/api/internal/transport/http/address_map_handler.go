package httptransport

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/neshan"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ReverseGeocoder interface {
	Lookup(context.Context, float64, float64) (neshan.Result, error)
}
type mapUsage struct {
	day   string
	total int
	users map[string][]time.Time
}
type AddressMapHandler struct {
	service *backoffice.Service
	auth    *auth.Service
	webKey  string
	reverse ReverseGeocoder
	limit   int
	mu      sync.Mutex
	usage   mapUsage
}

func NewAddressMapHandler(s *backoffice.Service, a *auth.Service, webKey string, reverse ReverseGeocoder, dailyLimit int) *AddressMapHandler {
	if dailyLimit < 1 {
		dailyLimit = 1000
	}
	return &AddressMapHandler{service: s, auth: a, webKey: webKey, reverse: reverse, limit: dailyLimit}
}
func (h *AddressMapHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Vary", "Authorization")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	session, e := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), identity.SurfaceCustomer)
	if e != nil {
		writeError(w, 401, "unauthorized")
		return
	}
	enabled := h.webKey != "" && h.reverse != nil
	if r.Method == "GET" {
		key := ""
		if enabled {
			key = h.webKey
		}
		writeJSON(w, 200, map[string]any{"enabled": enabled, "provider": "neshan", "web_key": key})
		return
	}
	var in struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	if in.Latitude == nil || in.Longitude == nil || math.IsNaN(*in.Latitude) || math.IsNaN(*in.Longitude) || math.IsInf(*in.Latitude, 0) || math.IsInf(*in.Longitude, 0) || *in.Latitude < -90 || *in.Latitude > 90 || *in.Longitude < -180 || *in.Longitude > 180 {
		writeError(w, 400, "invalid_request")
		return
	}
	if !enabled {
		writeError(w, 503, "map_not_configured")
		return
	}
	now := time.Now().UTC()
	h.mu.Lock()
	day := now.Format("2006-01-02")
	if h.usage.day != day {
		h.usage = mapUsage{day: day, users: map[string][]time.Time{}}
	}
	attempts := h.usage.users[session.User.ID][:0]
	for _, v := range h.usage.users[session.User.ID] {
		if now.Sub(v) < time.Hour {
			attempts = append(attempts, v)
		}
	}
	if len(attempts) >= 60 || h.usage.total >= h.limit {
		h.mu.Unlock()
		writeError(w, 429, "map_rate_limited")
		return
	}
	h.usage.total++
	h.usage.users[session.User.ID] = append(attempts, now)
	h.mu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result, e := h.reverse.Lookup(ctx, *in.Latitude, *in.Longitude)
	if e != nil {
		writeError(w, 503, "map_provider_unavailable")
		return
	}
	config, e := h.service.PublicCommerceConfig(r.Context())
	if e != nil {
		writeError(w, 503, "map_provider_unavailable")
		return
	}
	province, city := "", ""
	norm := func(s string) string {
		s = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "ي", "ی"), "ك", "ک"))
		return strings.Join(strings.Fields(strings.TrimPrefix(s, "استان ")), " ")
	}
	if directory, ok := config["geography"].([]commerce.Province); ok {
		for _, p := range directory {
			if !p.Active || norm(p.Name) != norm(result.Province) {
				continue
			}
			for _, c := range p.Cities {
				if c.Active && norm(c.Name) == norm(result.City) {
					province, city = p.Name, c.Name
				}
			}
		}
	}
	writeJSON(w, 200, map[string]any{"full_address": result.Address, "province": province, "city": city, "geography_matched": province != "" && city != ""})
}
