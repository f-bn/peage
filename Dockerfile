FROM docker.io/golang:1.24.6 AS build

WORKDIR /build

COPY . .

RUN go build -ldflags="-s -w" -o peage

FROM cgr.dev/chainguard/wolfi-base:latest

COPY --from=build --chown=0:0 --chmod=0755 \
  /build/peage /usr/bin/peage

ENTRYPOINT [ "/usr/bin/peage" ]

EXPOSE 2375/tcp

LABEL \
  org.opencontainers.image.title="peage" \
  org.opencontainers.image.source="https://github.com/f-bn/peage" \
  org.opencontainers.image.version="0.1.0" \
  org.opencontainers.image.description="Simple Docker API socket filtering reverse proxy" \
  org.opencontainers.image.licenses="BSD-3-Clause" \
  org.opencontainers.image.authors="Florian Bobin <contact@fbobin.me>"