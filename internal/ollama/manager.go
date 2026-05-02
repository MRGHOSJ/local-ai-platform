package ollama

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	apiURL      = "http://127.0.0.1:11434"
	ollamaPort  = "127.0.0.1:11434"
)

var (
	manager *Manager
)

type Manager struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	ctx    context.Context
}

func GetManager() *Manager {
	if manager == nil {
		manager = &Manager{}
	}
	return manager
}

func (m *Manager) Start() error {
	if m.IsRunning() {
		return nil
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())

	binaryPath, err := m.findOllamaBinary()
	if err != nil {
		return fmt.Errorf("ollama binary not found: %w", err)
	}

	if err := os.MkdirAll(getModelsDir(), 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	if err := os.MkdirAll(getTempDir(), 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	env := os.Environ()
	env = append(env, "OLLAMA_HOST="+ollamaPort)
	env = append(env, "OLLAMA_MODELS="+getModelsDir())
	env = append(env, "OLLAMA_TMPDIR="+getTempDir())

	m.cmd = exec.CommandContext(m.ctx, binaryPath, "serve")
	m.cmd.Env = env
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		m.cancel()
		return fmt.Errorf("failed to start ollama: %w", err)
	}

	return m.waitForReady(60 * time.Second)
}

func (m *Manager) Stop() error {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.cmd != nil && m.cmd.Process != nil {
		return m.cmd.Process.Kill()
	}
	return nil
}

func (m *Manager) IsRunning() bool {
	resp, err := http.Get(apiURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (m *Manager) waitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if m.IsRunning() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("ollama failed to start within %v", timeout)
}

func (m *Manager) findOllamaBinary() (string, error) {
	// Use installer to find/download ollama (System → Cache → Download)
	path, err := EnsureOllama(nil)
	if err != nil {
		return "", err
	}
	return path, nil
}

func getAppDir() string {
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "local-ai-platform")
}

func getModelsDir() string {
	return filepath.Join(getAppDir(), "models")
}

func getTempDir() string {
	return filepath.Join(getAppDir(), "tmp")
}

func GetAPIURL() string {
	return apiURL
}

func GetModelsDirPath() string {
	return getModelsDir()
}