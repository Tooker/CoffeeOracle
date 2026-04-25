package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// RouterOptions holds dependencies for HTTP handlers.
type RouterOptions struct {
	OracleHandler http.Handler
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
	return mux
}

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := healthResponse{Status: "ok", Timestamp: time.Now().UTC()}
	_ = json.NewEncoder(w).Encode(resp)
}
