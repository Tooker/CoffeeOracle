package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// RouterOptions holds dependencies for HTTP handlers.
type RouterOptions struct {
	OracleHandler http.Handler
	ImageHandler  http.Handler
	ShareHandler  http.Handler
	Middleware    []func(http.Handler) http.Handler
}

// NewRouter configures the HTTP routes exposed by the backend service.
func NewRouter(opts RouterOptions) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)
	if opts.OracleHandler != nil {
		h := opts.OracleHandler
		for i := len(opts.Middleware) - 1; i >= 0; i-- {
			h = opts.Middleware[i](h)
		}
		mux.Handle("/api/oracle", h)
	}
	if opts.ImageHandler != nil {
		mux.Handle("/api/image/", opts.ImageHandler)
	}
	if opts.ShareHandler != nil {
		mux.Handle("/api/share/", opts.ShareHandler)
	}
	return mux
}

// healthResponse is the payload returned by /healthz for liveness checks.
type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// healthzHandler answers monitoring probes with a minimal JSON status payload.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := healthResponse{Status: "ok", Timestamp: time.Now().UTC()}
	_ = json.NewEncoder(w).Encode(resp)
}
