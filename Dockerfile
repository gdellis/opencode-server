# Build stage for auth proxy
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-proxy ./cmd/server

# Runtime stage - extend official opencode image
FROM ghcr.io/anomalyco/opencode:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy auth proxy binary
COPY --from=builder /app/auth-proxy /auth-proxy
COPY --from=builder /app/static /static

# Create projects directory
RUN mkdir -p /var/opencode/projects

EXPOSE 4096

ENTRYPOINT ["/auth-proxy"]