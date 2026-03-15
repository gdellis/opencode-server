# Testing Guidelines

## Test File Placement

Place `*_test.go` files alongside the source files they test:

```
internal/auth/auth.go
internal/auth/auth_test.go

internal/project/project.go
internal/project/project_test.go
```

## HTTP Handler Testing

Use `httptest` for testing HTTP handlers:

```go
func TestLoginHandler(t *testing.T) {
    // Create request
    body := strings.NewReader("password=testpassword")
    req := httptest.NewRequest("POST", "/login", body)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    // Create response recorder
    w := httptest.NewRecorder()
    
    // Call handler
    handler(w, req)
    
    // Check response
    if w.Code != http.StatusFound {
        t.Errorf("expected status 302, got %d", w.Code)
    }
    
    // Check redirect location
    location := w.Header().Get("Location")
    if location != "/" {
        t.Errorf("expected redirect to /, got %s", location)
    }
    
    // Check cookie was set
    cookies := w.Result().Cookies()
    if len(cookies) == 0 {
        t.Error("expected session cookie to be set")
    }
}
```

## Testing Protected Endpoints

```go
func TestProtectedEndpoint(t *testing.T) {
    // Create middleware with test credentials
    middleware := auth.NewMiddleware(testHash, testSecret, time.Hour, 5, time.Minute)
    
    // Create protected handler
    handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    
    // Test without auth
    req := httptest.NewRequest("GET", "/api/projects", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    if w.Code != http.StatusFound {
        t.Errorf("expected redirect without auth, got %d", w.Code)
    }
    
    // Test with valid session
    req = httptest.NewRequest("GET", "/api/projects", nil)
    req.AddCookie(&http.Cookie{
        Name:  "opencode_session",
        Value: validSessionToken,
    })
    w = httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected success with auth, got %d", w.Code)
    }
}
```

## Mocking Interfaces

Use interfaces for dependency injection in tests:

```go
// Define interface in production code
type ProcessRestarter interface {
    Restart(projectsDir string) error
}

// Mock in test code
type MockProcess struct {
    RestartCalled bool
    RestartPath   string
    RestartError  error
}

func (m *MockProcess) Restart(projectsDir string) error {
    m.RestartCalled = true
    m.RestartPath = projectsDir
    return m.RestartError
}

// Use in test
func TestSwitchProject(t *testing.T) {
    mock := &MockProcess{}
    manager := project.NewManagerWithProcess("/tmp/projects", mock)
    
    // Create test project
    os.MkdirAll("/tmp/projects/testproject", 0755)
    
    // Test switch
    // ...
    
    if !mock.RestartCalled {
        t.Error("expected Restart to be called")
    }
}
```

## Table-Driven Tests

```go
func TestValidateProjectName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid lowercase", "my-project", false},
        {"valid uppercase", "My-Project", false},
        {"valid numbers", "project-123", false},
        {"valid underscore", "my_project", false},
        {"empty", "", true},
        {"with spaces", "my project", true},
        {"with special char", "my-project!", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := process.ValidateProjectName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
        })
    }
}
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./internal/auth

# Run single test
go test -v -run TestLoginHandler ./internal/auth

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```