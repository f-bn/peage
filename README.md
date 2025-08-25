## Overview

:warning: **This is an early work and my first Go project, issues can arise** :warning:

Peage is a small program written in Go that filters calls to the Docker or Podman API when the UNIX socket is used.

This allows some softwares such as Traefik or Prometheus to use their built-in service discovery mechanisms a bit more safely by running in non-privileged mode (i.e removing the need of mounting the Docker/Podman socket into the container and therefore running as root).

The goal of this software is to remain as simple as possible by not covering all possible use cases for a filtering reverse proxy (prefer alternative solutions if needed).

## Usage

```
  -engine string
        Container engine API used for filtering (must be 'docker' or 'podman') (default "docker")
  -listen-addr string
        Listen address for the Peage reverse proxy server (default "localhost:2375")
  -socket string
        Path to the container engine API UNIX socket (default "/var/run/docker.sock")
  -verbose
        Enable verbose logging of requests
```

The easiest way to use Peage is to use the container image:

```shell
docker run -d -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/f-bn/peage:0.4.0 [flags]
```

### Allowed endpoints

Peage only allows calls using the `GET` or `HEAD` method on specific **hardcoded** paths depending of the choosen engine:

**Docker**

  - `/containers/json`
  - `/containers/*/json`
  - `/events`
  - `/images/json`
  - `/images/[^/]+/json`
  - `/info`
  - `/networks`
  - `/version`
  - `/_ping`

**Podman**

  - `/libpod/containers/json`
  - `/libpod/containers/stats`
  - `/libpod/containers/*/(json|changes|exists|stats)`
  - `/libpod/events`
  - `/libpod/images/json`
  - `/libpod/images/*/(json|exists)`
  - `/libpod/info`
  - `/libpod/networks/json`
  - `/libpod/networks/*/(json|exists)`
  - `/libpod/pods/json`
  - `/libpod/pods/stats`
  - `/libpod/pods/*/(json|exists)`
  - `/libpod/_ping`
  - `/libpod/version`
  - `/libpod/volumes/json`
  - `/libpod/volumes/*/(json|exists)`

Note: Podman API filtering is only done on libpod 5.0.0+ API paths (Docker-compatible API paths are not allowed, use the `docker` filtering mode for this)

## Compatibility

Peage is compatible with any software implementing the Docker or Pdoamn API spec.

I use it personally in front of the Podman API (through the Docker-compatible API endpoint).
