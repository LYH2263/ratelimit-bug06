package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/LYH2263/go-ratelimit/ratelimit"
)

type API struct {
	Limiter *ratelimit.Limiter
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/stats":
		a.stats(w, r)
	case "/api/keys":
		a.keys(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *API) stats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, a.Limiter.Stats())
}

func (a *API) keys(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, a.Limiter.SnapshotKeys())
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
