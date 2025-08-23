package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
)

// Flags
var (
	listenAddress string
	socketPath    string
)

var allowedPaths = []string{
	"^/containers/json$",
	"^/containers/[^/]+/json$",
	"^/events$",
	"^/info$",
	"^/networks$",
	"^/version$",
	"^/_ping$",
}

func init() {
	flag.StringVar(&listenAddress, "listen-addr", "localhost:2375", "Listen address for the Peage HTTP server")
	flag.StringVar(&socketPath, "socket", "/var/run/docker.sock", "Path to the Docker API UNIX socket")
}

func isAllowedPath(path string) bool {
	versionPattern := `^/v\d+\.\d+`
	if match, _ := regexp.MatchString(versionPattern, path); match {
		re := regexp.MustCompile(versionPattern)
		path = re.ReplaceAllString(path, "")
	}

	for _, p := range allowedPaths {
		if match, _ := regexp.MatchString(p, path); match {
			return true
		}
	}
	return false
}

func NewDockerProxy() *httputil.ReverseProxy {
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

func proxyHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.UserAgent()

		// Only allows GET method
		if r.Method != http.MethodGet {
			slog.Info("Blocked invalid request: non-allowed method", "method", r.Method, "path", r.URL.Path, "client", userAgent)
			http.Error(w, "Invalid request: method not allowed (supported method: GET)", http.StatusMethodNotAllowed)
			return
		}

		// Check if the path is allowed
		if !isAllowedPath(r.URL.Path) {
			slog.Info("Blocked invalid request: non-allowed path", "method", r.Method, "path", r.URL.Path, "client", userAgent)
			http.Error(w, "Invalid request: path not allowed", http.StatusForbidden)
			return
		}

		slog.Info("Forwarded valid request", "method", r.Method, "path", r.URL.Path, "client", userAgent)
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	// Flags parsing
	flag.Parse()

	// Preflight checks
	if _, err := os.Stat(socketPath); err != nil {
		slog.Error("Docker API UNIX socket not found, is Docker running?", "error", err)
		os.Exit(1)
	}
	slog.Info("Docker API UNIX socket found", "path", socketPath)

	// Create the reverse proxy
	proxy := NewDockerProxy()

	// Register proxy handler
	http.HandleFunc("/", proxyHandler(proxy))

	slog.Info("Starting Peage server", "address", listenAddress)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
