FROM docker.io/golang:1.26.4 AS build

ARG VERSION="0.6.0"

WORKDIR /build

ADD . .

RUN go build -ldflags="-s -w \
    -X main.version=${VERSION} \
    -X main.commitHash=$(git rev-parse HEAD | cut -c1-8) \
    -X main.buildDate=$(date -u '+%Y-%m-%d_%I:%M:%S%p')" \
  -o peage ./cmd/peage/

FROM cgr.dev/chainguard/wolfi-base:latest

ARG VERSION="0.6.0"

COPY --from=build --chown=0:0 --chmod=0755 \
  /build/peage /usr/bin/peage

ENTRYPOINT [ "/usr/bin/peage" ]

EXPOSE 2375/tcp

LABEL \
  org.opencontainers.image.title="peage" \
  org.opencontainers.image.source="https://github.com/f-bn/peage" \
  org.opencontainers.image.version="${VERSION}" \
  org.opencontainers.image.description="Simple container engine API socket filtering reverse proxy" \
  org.opencontainers.image.licenses="BSD-3-Clause" \
  org.opencontainers.image.authors="Florian Bobin <contact@fbobin.me>"
