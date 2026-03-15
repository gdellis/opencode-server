# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-proxy ./cmd/server

# Runtime stage
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/auth-proxy /auth-proxy
COPY --from=builder /app/static /static

# Create projects directory
RUN mkdir -p /var/opencode/projects && chown appuser:appuser /var/opencode/projects

# Switch to non-root user
USER appuser

EXPOSE 4096

ENTRYPOINT ["/auth-proxy"]