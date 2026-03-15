# AGENTS.md

Coding agent instructions for the OpenCode Remote Server project.

## Project Overview

Go-based authentication wrapper for OpenCode's web interface. Extends `ghcr.io/anomalyco/opencode:latest` Docker image.

**Tech Stack:** Go 1.24, Docker, bcrypt authentication, reverse proxy, project management

**Key Features:**
- Custom login UI with bcrypt authentication
- Session management with HTTP-only cookies
- Project management (create/switch projects)
- Rate limiting (5 attempts/minute by IP)
- Reverse proxy to OpenCode server

## Documentation

| File | Content |
|------|---------|
| [docs/BUILD.md](docs/BUILD.md) | Build, lint, test commands |
| [docs/CODE_STYLE.md](docs/CODE_STYLE.md) | Detailed code style guidelines |
| [docs/TESTING.md](docs/TESTING.md) | Testing patterns and conventions |

## Quick Start

```bash
make setup                               # Generate .env (first-time)
make up                                  # Start containers
make logs                                # View logs
make down                                # Stop containers
```

## Project Structure

```
cmd/server/main.go           # Entry point, config
internal/auth/auth.go         # Auth middleware, login UI
internal/process/process.go   # OpenCode process management
internal/project/project.go   # Project CRUD
Dockerfile                    # Extends ghcr.io/anomalyco/opencode:latest
docker-compose.yml
Makefile
setup.sh                      # First-time setup script
```

## Key Conventions

### Configuration
Environment variables with defaults in `loadConfig()`:

```go
ListenAddr:  getEnv("LISTEN_ADDR", ":4096"),
ProjectsDir: getEnv("PROJECTS_DIR", "/var/opencode/projects"),
```

### Constructors
Use `New<Type>` pattern:

```go
func NewManager(projectsDir string) *Manager { ... }
func NewMiddleware(...) *Middleware { ... }
```

### Error Handling
Wrap errors with context using `%w`:

```go
return fmt.Errorf("failed to create directory: %w", err)
```

### Concurrency
Use `sync.RWMutex` for shared state. RLock for reads, Lock for writes.

### HTTP Handlers
Use `http.ServeMux` with method-specific patterns:

```go
mux.HandleFunc("GET /api/projects", handler)
mux.HandleFunc("POST /api/projects", handler)
```

## Security Notes

- **Passwords**: bcrypt cost 10, must use `$2a$` format
- **.env escaping**: Escape `$` as `$$` in bcrypt hash for Docker Compose
- **Sessions**: 64-char hex tokens, HTTP-only, Secure, SameSite=Strict
- **Rate limiting**: 5 attempts/minute by IP, resets after 1 minute
- **Input validation**: alphanumeric, hyphen, underscore for project names

## Common Tasks

### Add new endpoint
1. Add handler in `internal/<package>/<package>.go`
2. Register route in `cmd/server/main.go`
3. Add auth middleware if protected

### Change port
Edit `docker-compose.yml`: `- "8080:4096"`

### Debug authentication
Check hash format (must be `$2a$`), rate limit resets after 1 minute.

### Add dependency
```bash
go get golang.org/x/sync@latest
go mod tidy
```

## File References

- Setup script: `setup.sh` - Interactive password/environment setup
- Makefile: Build and container commands
- README.md: User documentation