# syntax=docker/dockerfile:1.7

# Runtime-only image. The binary is built by goreleaser (see
# .goreleaser.yaml dockers:) and dropped into the build context as
# `chronos`; this file just packages it on top of distroless.
#
# CGO is disabled in the goreleaser build because Chronos uses the
# pure-Go modernc.org/sqlite driver, so the binary runs on any
# Linux/amd64 or arm64 base image without a libc dependency.
#
# For local Docker builds (no goreleaser), `make docker-build` runs
# `make build` first to populate ./bin/chronos and re-uses this same
# Dockerfile via a copy step in the Makefile.

FROM gcr.io/distroless/static:nonroot@sha256:e3f945647ffb95b5839c07038d64f9811adf17308b9121d8a2b87b6a22a80a39

LABEL org.opencontainers.image.title="chronos" \
      org.opencontainers.image.description="Time / Pattern Perception in the cognitive stack" \
      org.opencontainers.image.source="https://github.com/felixgeelhaar/chronos" \
      org.opencontainers.image.licenses="MIT"

COPY chronos /chronos

# The HTTP server defaults to :7778. Bind 0.0.0.0 in containerised
# deployments via CHRONOS_HTTP_HOST=0.0.0.0.
EXPOSE 7778

# Default to the server; override with `chronos compute ...` when
# running as a job.
ENTRYPOINT ["/chronos"]
CMD ["serve"]
