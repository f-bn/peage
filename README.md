## Overview

:warning: **This is an early work and my first Go project, issues can arise** :warning:

Peage is a small program written in Go that filters calls to the Docker or Podman API when the UNIX socket is used.

This allows some softwares such as Traefik or Prometheus to use their built-in service discovery mechanisms a bit more safely by running in non-privileged mode (i.e removing the need of mounting the Docker/Podman socket into the container and therefore running as root).

The goal of this software is to remain as simple as possible by not covering all possible use cases for a filtering reverse proxy (prefer alternative solutions if needed).

## Usage

```
  -engine string
        Container engine API used for filtering (must be 'docker', 'podman' or 'podman-compat') (default "docker")
  -listen-addr string
        Listen address for the Peage reverse proxy server (default "localhost:2375")
  -socket string
        Path to the container engine API UNIX socket (default "/var/run/docker.sock")
  -verbose
        Enable verbose logging of requests
```

The easiest way to use Peage is to use the container image:

```console
$ docker run -d --name peage \
  -p 2375:2375 -v /var/run/docker.sock:/var/run/docker.sock:ro \
  ghcr.io/f-bn/peage:0.5.0 \
    --listen-addr=:2375\
    --verbose
```

Then, you can send your request (i.e with cURL):

```console
$ curl http://localhost:2375/v1.47/_ping
OK
```

The request has been forwarded successfuly as it match one the allowed endpoints:

```console
$ docker logs peage
time=2025-08-27T15:24:13.899Z level=INFO msg="Starting Peage" version=0.5.0 commit=b19ae059 buildDate=2025-08-27_03:19:57PM
time=2025-08-27T15:24:13.899Z level=INFO msg="Preflight checks passed"
time=2025-08-27T15:24:13.899Z level=INFO msg="Container engine API socket found" engine=docker path=/var/run/docker.sock
time=2025-08-27T15:24:13.899Z level=INFO msg="Starting reverse proxy" address=:2375
time=2025-08-27T15:24:21.914Z level=DEBUG msg="Forwarded valid request" method=GET path=/v1.47/_ping client=curl/8.12.1
```

Same goes for Podman API, you need to set some flags to correctly target the Podman API socket:

```console
$ podman run -d --name peage \
  -p 2375:2375 -v /run/podman/podman.sock:/run/podman/podman.sock:ro \
  ghcr.io/f-bn/peage:0.5.0 \
    --listen-addr=:2375 \
    --engine=podman \
    --socket=/run/podman/podman.sock \
    --verbose

$ curl http://localhost:2375/v5.5.2/libpod/_ping
OK

$ podman logs peage
time=2025-08-27T15:27:13.341Z level=INFO msg="Starting Peage" version=0.5.0 commit=b19ae059 buildDate=2025-08-27_03:19:57PM
time=2025-08-27T15:27:13.341Z level=INFO msg="Preflight checks passed"
time=2025-08-27T15:27:13.341Z level=INFO msg="Container engine API socket found" engine=podman path=/run/podman/podman.sock
time=2025-08-27T15:27:13.341Z level=INFO msg="Starting reverse proxy" address=:2375
time=2025-08-27T15:27:35.688Z level=DEBUG msg="Forwarded valid request" method=GET path=/v5.5.2/libpod/_ping client=curl/8.12.1
```

### Allowed endpoints

Peage only allows calls using the `GET` or `HEAD` method on specific **hardcoded** paths depending of the choosen engine filtering mode:

**Docker (docker)**

  - `/containers/json`
  - `/containers/*/json`
  - `/events`
  - `/images/json`
  - `/images/*/json`
  - `/info`
  - `/networks`
  - `/version`
  - `/volumes`
  - `/volumes/<name>`
  - `/_ping`

**Podman (podman)**

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

**Podman with Docker-compatible endpoints (podman-compat)**

This engine mode enable both Docker and Podman API filtering. This is useful if you want to have a single proxy on top of Podman API to handle both Docker-compatible and dedicated Podman endpoints.

For example, if you have apps that only know about Docker API (i.e Traefik) and some others only about Podman API (i.e Prometheus Podman Exporter), then it is easier to manage with a single proxy instead of having to deploy one for each filtering mode.

## Compatibility

Peage is compatible with any software implementing the Docker or Podman API spec.
