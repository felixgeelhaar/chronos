# syntax=docker/dockerfile:1.7

# Multi-stage build producing a static, distroless chronos binary.
# CGO is disabled because we use the pure-Go modernc.org/sqlite driver,
# so the resulting binary runs on any Linux/amd64 or arm64 base image
# without a libc dependency.

FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache module downloads independently from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build args wired from CI / Make to inject version metadata.
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown
ARG TARGETOS=linux
ARG TARGETARCH=amd64

ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH}

RUN go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /chronos \
    ./cmd/chronos

# Distroless static gives us a minimal runtime image (~2MB) with a
# non-root user, /etc/passwd, and ca-certificates baked in. No shell.
FROM gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.title="chronos" \
      org.opencontainers.image.description="Time / Pattern Perception in the cognitive stack" \
      org.opencontainers.image.source="https://github.com/felixgeelhaar/chronos" \
      org.opencontainers.image.licenses="MIT"

COPY --from=builder /chronos /chronos

# The HTTP server defaults to :7778. Bind 0.0.0.0 in containerised
# deployments via CHRONOS_HTTP_HOST=0.0.0.0.
EXPOSE 7778

# Default to the server; override with `chronos compute ...` when
# running as a job.
ENTRYPOINT ["/chronos"]
CMD ["serve"]
