package process

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const DefaultOpenCodePath = "/usr/local/bin/opencode"

type OpenCode struct {
	mu          sync.RWMutex
	cmd         *exec.Cmd
	path        string
	projectsDir string
	addr        string
	args        []string
	proxy       *httputil.ReverseProxy
}

func NewOpenCode(path, projectsDir, addr string, args []string) *OpenCode {
	if path == "" {
		path = DefaultOpenCodePath
	}
	target, _ := url.Parse("http://" + addr)
	return &OpenCode{
		path:        path,
		projectsDir: projectsDir,
		addr:        addr,
		args:        args,
		proxy:       httputil.NewSingleHostReverseProxy(target),
	}
}

func (o *OpenCode) Start() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cmd != nil && o.cmd.Process != nil {
		return nil
	}

	if err := os.MkdirAll(o.projectsDir, 0755); err != nil {
		return fmt.Errorf("failed to create projects directory: %w", err)
	}

	args := []string{"serve", "--hostname", o.addr}
	args = append(args, o.args...)

	o.cmd = exec.Command(o.path, args...)
	o.cmd.Dir = o.projectsDir
	o.cmd.Stdout = os.Stdout
	o.cmd.Stderr = os.Stderr

	if err := o.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start opencode: %w", err)
	}

	return nil
}

func (o *OpenCode) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cmd != nil && o.cmd.Process != nil {
		o.cmd.Process.Signal(os.Interrupt)
		o.cmd.Wait()
		o.cmd = nil
	}
}

func (o *OpenCode) Restart(projectsDir string) error {
	o.Stop()
	o.projectsDir = projectsDir
	return o.Start()
}

func (o *OpenCode) Proxy() http.Handler {
	return o.proxy
}

func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("project name can only contain letters, numbers, hyphens, and underscores")
		}
	}
	return nil
}

func IsProject(path, name string) bool {
	stat, err := os.Stat(filepath.Join(path, name))
	return err == nil && stat.IsDir()
}
