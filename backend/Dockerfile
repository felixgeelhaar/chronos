# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY migrations/ ./migrations/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -o api ./cmd/api

# Stage 2: Production
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S ascend && \
    adduser -S ascend -u 1001 -G ascend

# Copy binary and migrations from builder
COPY --from=builder --chown=ascend:ascend /build/api ./api
COPY --from=builder --chown=ascend:ascend /build/migrations ./migrations

# Set timezone to UTC
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Switch to non-root user
USER ascend

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Start application
CMD ["./api"]
