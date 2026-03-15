# Code Style Guidelines

## Imports

Group imports: standard library first, then external packages.

```go
import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "sync"
    
    "github.com/example/opencode-server/internal/auth"
    "golang.org/x/crypto/bcrypt"
)
```

## Types

### Exported Structs

PascalCase with JSON tags for exported structs:

```go
type Project struct {
    Name    string `json:"name"`
    Path    string `json:"path"`
    Current bool   `json:"current"`
}

type Config struct {
    ListenAddr    string
    InternalAddr  string
    SessionExpiry time.Duration
}
```

### Interfaces

Define interfaces for dependency injection:

```go
type ProcessRestarter interface {
    Restart(projectsDir string) error
}
```

## Naming Conventions

| Element | Convention | Example |
|---------|------------|----------|
| Packages | short, lowercase | `auth`, `process`, `project` |
| Types | PascalCase | `Manager`, `Middleware`, `OpenCode` |
| Methods | PascalCase (exported) | `NewManager`, `ListProjects` |
| Private fields | camelCase | `projectsDir`, `currentProj` |
| Receivers | 1-2 letters | `m *Manager`, `o *OpenCode` |
| Constants | package level | `const cookieName = "opencode_session"` |

## Error Handling

### Wrap with Context

Use `fmt.Errorf` with `%w` for wrapping:

```go
if err := os.MkdirAll(dir, 0755); err != nil {
    return fmt.Errorf("failed to create directory: %w", err)
}
```

### Fatal Errors in Main

Use `log.Fatalf` for startup failures:

```go
if err := opencodeProc.Start(); err != nil {
    log.Fatalf("Failed to start opencode: %v", err)
}
```

### HTTP Error Responses

```go
http.Error(w, "Invalid request body", http.StatusBadRequest)
http.Error(w, "Project not found", http.StatusNotFound)
http.Error(w, "Internal server error", http.StatusInternalServerError)
```

## Concurrency

### Mutex Protection

Use `sync.RWMutex` for shared state. RLock for reads, Lock for writes:

```go
type Manager struct {
    mu          sync.RWMutex
    projectsDir string
    currentProj string
}

// Read operation
func (m *Manager) ListProjects() []Project {
    m.mu.RLock()
    defer m.mu.RUnlock()
    // ...
}

// Write operation
func (m *Manager) CreateProject() {
    m.mu.Lock()
    defer m.mu.Unlock()
    // ...
}
```

## HTTP Handlers

### Route Registration

Use `http.ServeMux` with method patterns:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /api/projects", projManager.ListProjects)
mux.HandleFunc("POST /api/projects", projManager.CreateProject)
```

### Middleware Pattern

Wrap handlers for authentication:

```go
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie(cookieName)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        if !m.validateSession(cookie.Value) {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## Configuration

### Environment Variables with Defaults

```go
func loadConfig() *Config {
    return &Config{
        ListenAddr:    getEnv("LISTEN_ADDR", ":4096"),
        InternalAddr:  getEnv("INTERNAL_ADDR", "127.0.0.1:4097"),
        ProjectsDir:   getEnv("PROJECTS_DIR", "/var/opencode/projects"),
    }
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
```

## Constants

Define at package level:

```go
const cookieName = "opencode_session"
const DefaultOpenCodePath = "/usr/local/bin/opencode"
```

## Constructor Functions

Use `New<Type>` pattern:

```go
func NewMiddleware(passwordHash, sessionSecret string, ...) *Middleware {
    return &Middleware{
        passwordHash: []byte(passwordHash),
        // ...
    }
}

func NewManager(projectsDir string) *Manager {
    return &Manager{
        projectsDir: projectsDir,
    }
}
```

## Graceful Shutdown

Handle SIGINT/SIGTERM properly:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("Shutting down...")
opencodeProc.Stop()
mainServer.Close()
```