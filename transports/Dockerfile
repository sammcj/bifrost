# --- First Stage: Builder image ---
FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN echo $GOPATH
ENV GOPATH=/go
# Install dependencies in a single layer
RUN apk add --no-cache upx

# Set environment for static build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Define build-time variable for transport type
ARG TRANSPORT_TYPE=http

# Install Go binary directly
RUN go install \
        -ldflags="-w -s -extldflags '-static'" \
        -a -trimpath \
        github.com/maximhq/bifrost/transports/bifrost-${TRANSPORT_TYPE}@latest

# Move binary to app directory
RUN mv "$(go env GOPATH)/bin/linux_amd64/bifrost-${TRANSPORT_TYPE}" /app/main

# Compress binary with upx
RUN upx --best --lzma /app/main

# Verify build succeeded
RUN test -f /app/main || (echo "Build failed" && exit 1)

# --- Second Stage: Minimal runtime image ---
FROM alpine:3.19
WORKDIR /app

# Copy necessary files and create structure in one layer
COPY --from=builder /app/main .
RUN mkdir -p /app/config && \
    adduser -D -s /bin/sh appuser && \
    chown -R appuser:appuser /app
USER appuser

# Environment variables with defaults
ENV APP_PORT=8080 \
    APP_POOL_SIZE=300 \
    APP_DROP_EXCESS_REQUESTS=false \
    APP_PLUGINS="" \
    APP_PROMETHEUS_LABELS=""

EXPOSE 8080

# Direct entrypoint with environment variable expansion
ENTRYPOINT ["/bin/sh", "-c", "exec /app/main -config /app/config/config.json -port \"${APP_PORT}\" -pool-size \"${APP_POOL_SIZE}\" -drop-excess-requests \"${APP_DROP_EXCESS_REQUESTS}\" -plugins \"${APP_PLUGINS}\" -prometheus-labels \"${APP_PROMETHEUS_LABELS}\""]