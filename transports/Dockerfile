# --- First Stage: Builder image ---
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install dependencies in a single layer
RUN apk add --no-cache upx

# Set environment for static build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Define build-time variables
ARG TRANSPORT_TYPE=http
ARG TAG_VERSION=latest

# Initialize go module and get bifrost with specified tag
RUN go mod init bifrost-build && \
    go get github.com/maximhq/bifrost/transports/bifrost-${TRANSPORT_TYPE}@${TAG_VERSION}

# Build the binary locally
RUN go build \
        -ldflags="-w -s -extldflags '-static'" \
        -a -trimpath \
        -o /app/main \
        github.com/maximhq/bifrost/transports/bifrost-${TRANSPORT_TYPE}

# Compress binary with upx
RUN upx --best --lzma /app/main

# Verify build succeeded
RUN test -f /app/main || (echo "Build failed" && exit 1)

# --- Second Stage: Minimal runtime image ---
FROM alpine:3.19
WORKDIR /app

# Create data directory and set up user
COPY --from=builder /app/main .
COPY docker-entrypoint.sh .
RUN mkdir -p /app/data/logs && \
    adduser -D -s /bin/sh appuser && \
    chown -R appuser:appuser /app && \
    chmod +x /app/docker-entrypoint.sh
USER appuser

# Environment variables with defaults
ENV APP_PORT=8080 \
    APP_PLUGINS=""

# Declare volume for data persistence
VOLUME ["/app/data"]
EXPOSE 8080

# Health check for container status monitoring
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${APP_PORT}/metrics || exit 1

# Use entrypoint script that handles volume permissions
ENTRYPOINT ["/bin/sh", "-c", "/app/docker-entrypoint.sh -app-dir /app/data -port ${APP_PORT} -plugins \"${APP_PLUGINS}\""]