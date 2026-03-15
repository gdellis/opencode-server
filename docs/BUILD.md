# Build Commands

## Local Build

```bash
go build ./cmd/server                    # Build binary
go build -ldflags="-s -w" -o auth-proxy ./cmd/server  # Optimized build
go run ./cmd/server                      # Run locally (requires opencode binary)
```

## Docker Build

```bash
docker compose build                     # Build container image
make build                               # Same as above
```

## Lint Quality

```bash
go vet ./...                             # Static analysis
go fmt ./...                             # Format code
go mod tidy                              # Clean dependencies
```

## Test Commands

```bash
go test ./...                            # Run all tests
go test -v ./internal/auth                # Run specific package
go test -run TestName ./...              # Run single test
go test -v -run TestName ./internal/auth  # Run single test verbose
```

## Container Operations

```bash
make setup                               # Generate .env (first-time)
make up                                  # Start containers
make down                                # Stop containers
make logs                                # View logs (follow)
make clean                               # Remove containers, volumes, .env
make restart                             # Rebuild and restart

docker compose exec opencode-server /bin/sh  # Shell into container
docker compose logs -f opencode-server       # Follow logs
```

## First-Time Setup

```bash
./setup.sh                               # Interactive setup script
# Prompts for:
# - Password (min 8 chars)
# - Session expiry (default: 24h)
# - Port (default: 4096)
# - Model (default: anthropic/claude-sonnet-4-5)
# - API keys (optional)
```