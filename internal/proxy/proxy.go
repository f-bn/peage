package proxy

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"

	"peage/internal/config"
)

func returnHTTPError(w http.ResponseWriter, errorCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func New(socketPath string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "localhost"
			req.Header.Set("Host", "peage")
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{}
				return dialer.DialContext(ctx, "unix", socketPath)
			},
		},
	}
}

func Handler(proxy *httputil.ReverseProxy, engine config.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.UserAgent()

		// Only allows GET or HEAD method
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			slog.Debug("Blocked invalid request: non-allowed method", "method", r.Method, "path", r.URL.Path, "client", userAgent)
			returnHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed (supported methods: GET, HEAD)")
			return
		}

		// Check if the path is allowed
		if !isAllowedPath(r.URL.Path, engine) {
			slog.Debug("Blocked invalid request: non-allowed path", "method", r.Method, "path", r.URL.Path, "client", userAgent)
			returnHTTPError(w, http.StatusForbidden, "Path not allowed")
			return
		}

		slog.Debug("Forwarded valid request", "method", r.Method, "path", r.URL.Path, "client", userAgent)
		proxy.ServeHTTP(w, r)
	}
}
