# Overview

Peage is a small program written in Go that filters calls to the Docker API when the UNIX socket is used.

The goal of this software is to remain as simple as possible by not covering all possible use cases for a filtering reverse proxy (prefer alternatives if needed).

Peage only allows calls using the `GET` method on the following hardcoded paths:

  - `/containers/json`
  - `/containers/*/json`,
  - `/events`
  - `/info`
  - `/networks`
  - `/version`
  - `/_ping`

This allows some softwares such as Traefik or Prometheus to use their built-in service discovery mechanisms a bit more safely by running in non-privileged mode (i.e no need to bind-mount Docker/Podman socket into the container).

# Usage

The easiest way to use Peage is to use the container image:

```shell
docker run -d -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/f-bn/peage:0.1.0
```

If you want to use it with Traefik for example:

```shell
# Start Traefik
docker run -d --name traefik \
  -p 80:80 \
  -p 443:443 \
  ghcr.io/f-bn/traefik:3.5.0 traefik \
    --log.level=DEBUG \
    --entrypoints.web.address=:80 \
    --entrypoints.websecure.address=:443 \
    --providers.docker \
    --providers.docker.endpoint=http://localhost:2375 \
    --providers.docker.exposedbydefault=false

# Start peage
# --net=container:peage is like running a pod, this allows the revese proxy
# to be only exposed in the Traefik container
docker run -d --name peage --net=container:traefik -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/f-bn/peage:0.1.0

# Run a demo container
docker run -d --name demo \
  --label "traefik.enable=true" \
  --label "traefik.http.routers.demo.rule=Host(\`demo.example.local\`)"
  docker.io/nginx:latest

# And voilà !
curl http://demo.example.local
...Welcome to nginx!...
```