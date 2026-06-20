package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type Engine string

const (
	EngineDocker       Engine = "docker"
	EnginePodman       Engine = "podman"
	EnginePodmanCompat Engine = "podman-compat"
)

type Config struct {
	ListenAddress string
	SocketPath    string
	Engine        Engine
	Verbose       bool
}

func Parse() (*Config, error) {
	cfg := &Config{}

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Simple Docker/Podman API socket filtering reverse proxy written in Go")
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n", flag.CommandLine.Name())
		flag.PrintDefaults()
	}

	flag.StringVar(&cfg.ListenAddress, "listen-addr", "localhost:2375", "Listen address for the Peage reverse proxy server")
	flag.StringVar(&cfg.SocketPath, "socket", "/var/run/docker.sock", "Path to the container engine API UNIX socket")
	flag.StringVar((*string)(&cfg.Engine), "engine", "docker", "Container engine API used for filtering (values: 'docker', 'podman', or 'podman-compat')")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging of requests")

	flag.Parse()

	if err := cfg.Engine.Validate(); err != nil {
		return nil, err
	}

	if err := checkSocketPathExists(cfg.SocketPath); err != nil {
		return nil, fmt.Errorf("socket path check failed: %w", err)
	}

	return cfg, nil
}

func (e Engine) Validate() error {
	switch e {
	case EngineDocker, EnginePodman, EnginePodmanCompat:
		return nil
	default:
		return errors.New("invalid container engine (must be 'docker', 'podman', or 'podman-compat')")
	}
}

func checkSocketPathExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}
