## Overview

Peage is a small program written in Go that filters calls to the Docker and Podman API when the UNIX socket is used.

This allows some softwares such as Traefik or Prometheus to use their built-in service discovery mechanisms based on those API a bit more safely by running in non-privileged mode. This remove the need of mounting the Docker/Podman socket into those containers and therefore to run them as root.

### Note regarding security

Even if this tool allows to avoid running some software that use the Docker/Podman API in non-privileged mode, this does **not** replace good security practices of deployment. It is still recommended to apply the principle of least privilege, keep your software up to date, and follow general container security guidelines.

Moreover, Peage **does not** aim to implement missing features in the Docker/Podman APIs such as authentication and authorization on top of them. It's sole purpose is to restrict which endpoints of those API are accessible, mostly for service discovery purposes (Kubernetes does a much better job on this part if that's needed, but this is not scope here).

## Usage

```
  -engine string
    	Container engine API used for filtering (values: 'docker', 'podman', or 'podman-compat') (default "docker")
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
  -p 127.0.0.1:2375:2375 -v /var/run/docker.sock:/var/run/docker.sock:ro \
  ghcr.io/f-bn/peage:0.6.0 \
    --listen-addr=:2375 \
    --verbose
```

> [!WARNING]
> On hosts with SELinux enabled, you need to disable label separation for the container otherwise Peage won't be able to forward requests to the API socket:
> * Docker: `--security-opt=label=disable`
> * Podman: `--security-opt=label=disabled`

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
  ghcr.io/f-bn/peage:0.6.0 \
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

- [docker](./internal/proxy/filter.go#L81)
- [podman](./internal/proxy/filter.go#L83)
- [podman-compat](./internal/proxy/filter.go#L85)

> [!NOTE]
> This `podman-compat` engine mode enable both Docker and Podman API filtering. This is useful if you want to have a single proxy on top of Podman API to handle both Docker-compatible and dedicated Podman endpoints.

## Compatibility

Peage is compatible with any software implementing the Docker or Podman API spec.

## Usage of AI disclosure

This project was reworked with the assistance of an AI coding agent as an engineering assistant to:

- Refactor and review Go code.
- Propose project structure, configurations, and testing improvements.
- Help debug issues.
- Generate and maintain documentation.

All AI-generated code was reviewed, tested, and validated by a human developer before being committed. The final decisions, design choices, and code quality remain the sole responsibility of the project maintainer.