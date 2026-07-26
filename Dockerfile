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

FROM gcr.io/distroless/static:nonroot@sha256:963fa6c544fe5ce420f1f54fb88b6fb01479f054c8056d0f74cc2c6000df5240

LABEL maintainer="Felix Geelhaar <felix.geelhaar@gmail.com>"

LABEL org.opencontainers.image.title="chronos" \
      org.opencontainers.image.description="Time / Pattern Perception in the cognitive stack" \
      org.opencontainers.image.source="https://github.com/felixgeelhaar/chronos" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.authors="Felix Geelhaar <felix.geelhaar@gmail.com>"

# root-owned and read-only-executable on purpose. Handing the binary to
# the runtime user would let a compromised process rewrite its own
# entrypoint and persist across a restart; 0555 under root:root leaves
# it executable by everyone and writable by no one.
COPY --chown=root:root --chmod=0555 chronos /chronos

# The HTTP server defaults to :7778. Bind 0.0.0.0 in containerised
# deployments via CHRONOS_HTTP_HOST=0.0.0.0.
EXPOSE 7778

# distroless:nonroot already defaults to uid 65532, but say it
# explicitly: the guarantee then survives a base-image retag, and it is
# visible to anyone reading this file or inspecting the image config.
USER 65532:65532

# The image has no shell and no curl, so the probe is the binary itself
# (`chronos health` -> GET /health on loopback). Exec form, because
# there is no /bin/sh to parse a shell-form command. /health is exempt
# from bearer auth, so this works with CHRONOS_API_TOKEN set.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/chronos", "health"]

# Default to the server; override with `chronos compute ...` when
# running as a job.
ENTRYPOINT ["/chronos"]
CMD ["serve"]
