package project

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/example/opencode-server/internal/process"
)

type ProcessRestarter interface {
	Restart(projectsDir string) error
}

type Manager struct {
	mu          sync.RWMutex
	projectsDir string
	currentProj string
}

type ManagerWithProcess struct {
	*Manager
	opencodeProc ProcessRestarter
}

func NewManager(projectsDir string) *Manager {
	return &Manager{
		projectsDir: projectsDir,
	}
}

func NewManagerWithProcess(projectsDir string, proc ProcessRestarter) *ManagerWithProcess {
	return &ManagerWithProcess{
		Manager:      NewManager(projectsDir),
		opencodeProc: proc,
	}
}

type Project struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Current bool   `json:"current"`
}

type CreateRequest struct {
	Name string `json:"name"`
}

type SwitchRequest struct {
	Name string `json:"name"`
}

func (m *Manager) ListProjects(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, err := os.ReadDir(m.projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]Project{})
			return
		}
		http.Error(w, "Failed to read projects directory", http.StatusInternalServerError)
		return
	}

	projects := make([]Project, 0)
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			projects = append(projects, Project{
				Name:    entry.Name(),
				Path:    filepath.Join(m.projectsDir, entry.Name()),
				Current: entry.Name() == m.currentProj,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(projects)
}

func (m *Manager) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := process.ValidateProjectName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if process.IsProject(m.projectsDir, req.Name) {
		http.Error(w, "Project already exists", http.StatusConflict)
		return
	}

	projectPath := filepath.Join(m.projectsDir, req.Name)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		http.Error(w, "Failed to create project directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Project{
		Name:    req.Name,
		Path:    projectPath,
		Current: false,
	})
}

func (m *ManagerWithProcess) SwitchProject(w http.ResponseWriter, r *http.Request) {
	var req SwitchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := process.ValidateProjectName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !process.IsProject(m.projectsDir, req.Name) {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	projectPath := filepath.Join(m.projectsDir, req.Name)

	if err := m.opencodeProc.Restart(projectPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to switch project: %v", err), http.StatusInternalServerError)
		return
	}

	m.currentProj = req.Name

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Project{
		Name:    req.Name,
		Path:    projectPath,
		Current: true,
	})
}
