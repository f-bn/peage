FROM cgr.dev/chainguard/wolfi-base:latest

ARG TARGETPLATFORM

ARG VERSION="dev"

COPY ${TARGETPLATFORM}/peage /usr/bin/peage

EXPOSE 2375/tcp

ENTRYPOINT [ "/usr/bin/peage" ]