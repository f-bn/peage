package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
)

// Version
var (
	version    = "dev"
	commitHash = "unknown"
	buildDate  = "unknown"
)

// Logging
var (
	logger   *slog.Logger
	logLevel slog.Level
)

// Flags
var (
	listenAddress   string
	socketPath      string
	containerEngine string
	verbose         bool
)

var dockerAllowedPaths = []string{
	"^/containers/json$",
	"^/containers/[^/]+/json$",
	"^/events$",
	"^/images/json$",
	"^/images/[^/]+/json$",
	"^/info$",
	"^/networks$",
	"^/version$",
	"^/volumes$",
	"^/volumes/[^/]+$",
	"^/_ping$",
}

var podmanAllowedPaths = []string{
	"^/libpod/containers/json$",
	"^/libpod/containers/stats$",
	"^/libpod/containers/[^/]+/(json|changes|exists|stats)$",
	"^/libpod/events$",
	"^/libpod/images/json$",
	"^/libpod/images/[^/]+/(json|exists)$",
	"^/libpod/info$",
	"^/libpod/networks/json$",
	"^/libpod/networks/[^/]+/(json|exists)$",
	"^/libpod/pods/json$",
	"^/libpod/pods/stats$",
	"^/libpod/pods/[^/]+/(json|exists)$",
	"^/libpod/_ping$",
	"^/libpod/version$",
	"^/libpod/volumes/json$",
	"^/libpod/volumes/[^/]+/(json|exists)$",
}

func init() {
	flag.StringVar(&listenAddress, "listen-addr", "localhost:2375", "Listen address for the Peage reverse proxy server")
	flag.StringVar(&socketPath, "socket", "/var/run/docker.sock", "Path to the container engine API UNIX socket")
	flag.StringVar(&containerEngine, "engine", "docker", "Container engine API used for filtering (must be 'docker' or 'podman')")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging of requests")
}

func runPreflightChecks() error {
	if !isValidContainerEngine(containerEngine) {
		return fmt.Errorf("invalid container engine '%s' (must be 'docker' or 'podman')", containerEngine)
	}
	if err := checkSocketPathExists(socketPath); err != nil {
		return fmt.Errorf("socket path check failed: %w", err)
	}
	return nil
}

func returnHTTPError(w http.ResponseWriter, errorCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func getEngineVersionPattern(engine string) string {
	switch engine {
	case "docker":
		return `^/v\d+\.\d+`
	case "podman":
		return `^/v\d+\.\d+(\.\d+)?`
	default:
		return ""
	}
}

func getEngineAllowedPaths(engine string) []string {
	switch engine {
	case "docker":
		return dockerAllowedPaths
	case "podman":
		return podmanAllowedPaths
	default:
		return nil
	}
}

func checkSocketPathExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func isValidContainerEngine(engine string) bool {
	return engine == "docker" || engine == "podman"
}

func isAllowedPath(path string) bool {
	versionPattern := getEngineVersionPattern(containerEngine)
	if versionPattern != "" {
		if match, _ := regexp.MatchString(versionPattern, path); match {
			re := regexp.MustCompile(versionPattern)
			path = re.ReplaceAllString(path, "")
		}
	}

	enginePaths := getEngineAllowedPaths(containerEngine)
	for _, p := range enginePaths {
		if match, _ := regexp.MatchString(p, path); match {
			return true
		}
	}
	return false
}

func NewSocketReverseProxy() *httputil.ReverseProxy {
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

		// Only allows GET or HEAD method
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			logger.Debug("Blocked invalid request: non-allowed method", "method", r.Method, "path", r.URL.Path, "client", userAgent)
			returnHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed (supported methods: GET, HEAD)")
			return
		}

		// Check if the path is allowed
		if !isAllowedPath(r.URL.Path) {
			logger.Debug("Blocked invalid request: non-allowed path", "method", r.Method, "path", r.URL.Path, "client", userAgent)
			returnHTTPError(w, http.StatusForbidden, "Path not allowed")
			return
		}

		logger.Debug("Forwarded valid request", "method", r.Method, "path", r.URL.Path, "client", userAgent)
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	// Flags parsing
	flag.Parse()

	// Logging
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	logger.Info("Starting Peage", "version", version, "commit", commitHash, "buildDate", buildDate)

	// Preflight checks
	if err := runPreflightChecks(); err != nil {
		logger.Error("Preflight checks failed", "error", err)
		os.Exit(1)
	}
	logger.Info("Container engine API socket found", "engine", containerEngine, "path", socketPath)

	// Create the reverse proxy
	proxy := NewSocketReverseProxy()

	// Register proxy handler
	http.HandleFunc("/", proxyHandler(proxy))

	logger.Info("Starting reverse proxy", "address", listenAddress)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
