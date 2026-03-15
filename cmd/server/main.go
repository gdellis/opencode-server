package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/example/opencode-server/internal/auth"
	"github.com/example/opencode-server/internal/process"
	"github.com/example/opencode-server/internal/project"
)

func main() {
	cfg := loadConfig()

	opencodeProc := process.NewOpenCode("", cfg.ProjectsDir, cfg.InternalAddr, cfg.OpenCodeArgs)
	projManager := project.NewManagerWithProcess(cfg.ProjectsDir, opencodeProc)

	authMiddleware := auth.NewMiddleware(
		cfg.PasswordHash,
		cfg.SessionSecret,
		cfg.SessionExpiry,
		5,
		time.Minute,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", authMiddleware.LoginPage)
	mux.HandleFunc("POST /login", authMiddleware.Login)
	mux.HandleFunc("GET /logout", authMiddleware.Logout)
	mux.HandleFunc("GET /api/health", healthHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/projects", projManager.ListProjects)
	protected.HandleFunc("POST /api/projects", projManager.CreateProject)
	protected.HandleFunc("POST /api/projects/switch", projManager.SwitchProject)
	protected.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyHandler(opencodeProc, w, r)
	})

	mainServer := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: authMiddleware.RequireAuth(protected),
	}

	log.Printf("Starting opencode process on %s", cfg.InternalAddr)
	if err := opencodeProc.Start(); err != nil {
		log.Fatalf("Failed to start opencode: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	go func() {
		log.Printf("Server listening on %s", cfg.ListenAddr)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	opencodeProc.Stop()
	mainServer.Close()
}

type Config struct {
	ListenAddr    string
	InternalAddr  string
	PasswordHash  string
	SessionSecret string
	SessionExpiry time.Duration
	ProjectsDir   string
	OpenCodeArgs  []string
}

func loadConfig() *Config {
	return &Config{
		ListenAddr:    getEnv("LISTEN_ADDR", ":4096"),
		InternalAddr:  getEnv("INTERNAL_ADDR", "127.0.0.1:4097"),
		PasswordHash:  os.Getenv("AUTH_PASSWORD_HASH"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		SessionExpiry: parseDuration(getEnv("SESSION_EXPIRY", "24h")),
		ProjectsDir:   getEnv("PROJECTS_DIR", "/var/opencode/projects"),
		OpenCodeArgs:  parseArgs(getEnv("OPENCODE_ARGS", "")),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func parseArgs(s string) []string {
	if s == "" {
		return nil
	}
	var args []string
	var current strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ' ' && !inQuote && current.Len() > 0:
			args = append(args, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"healthy": true})
}

func proxyHandler(opencodeProc *process.OpenCode, w http.ResponseWriter, r *http.Request) {
	opencodeProc.Proxy().ServeHTTP(w, r)
}
