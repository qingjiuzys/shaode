package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// EnvironmentManager manages shell environment state
type EnvironmentManager struct {
	mu            sync.RWMutex
	workingDir    string
	environment   map[string]string
	originalEnv   map[string]string // Original environment for restoration
}

// NewEnvironmentManager creates a new environment manager
func NewEnvironmentManager() *EnvironmentManager {
	em := &EnvironmentManager{
		environment: make(map[string]string),
		originalEnv: make(map[string]string),
	}

	// Store original environment
	em.initializeOriginalEnvironment()
	
	// Set initial working directory
	if wd, err := os.Getwd(); err == nil {
		em.workingDir = wd
	}

	return em
}

// initializeOriginalEnvironment stores the original environment variables
func (em *EnvironmentManager) initializeOriginalEnvironment() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, env := range os.Environ() {
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				key := env[:i]
				value := env[i+1:]
				em.originalEnv[key] = value
				em.environment[key] = value // Initialize with original values
				break
			}
		}
	}
}

// GetWorkingDir returns the current working directory
func (em *EnvironmentManager) GetWorkingDir() string {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.workingDir
}

// ChangeDir changes the current working directory
func (em *EnvironmentManager) ChangeDir(dir string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Handle relative paths
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(em.workingDir, dir)
	}

	// Clean the path
	dir = filepath.Clean(dir)

	// Check if directory exists
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	em.workingDir = dir
	return nil
}

// GetEnv gets an environment variable
func (em *EnvironmentManager) GetEnv(key string) string {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.environment[key]
}

// SetEnv sets an environment variable
func (em *EnvironmentManager) SetEnv(key, value string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.environment[key] = value
}

// UnsetEnv removes an environment variable
func (em *EnvironmentManager) UnsetEnv(key string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	delete(em.environment, key)
}

// GetAllEnv returns all environment variables
func (em *EnvironmentManager) GetAllEnv() map[string]string {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	// Return a copy to avoid concurrent modification
	envCopy := make(map[string]string)
	for k, v := range em.environment {
		envCopy[k] = v
	}
	return envCopy
}

// ExportEnvironment exports the current environment to the OS
func (em *EnvironmentManager) ExportEnvironment() {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Clear existing environment
	os.Clearenv()

	// Set new environment variables
	for key, value := range em.environment {
		os.Setenv(key, value)
	}
}

// RestoreOriginalEnvironment restores the original environment
func (em *EnvironmentManager) RestoreOriginalEnvironment() {
	em.mu.Lock()
	
	// Clear current environment
	em.environment = make(map[string]string)

	// Restore original values
	for key, value := range em.originalEnv {
		em.environment[key] = value
	}

	em.mu.Unlock()

	// Export to OS (without holding the lock)
	em.ExportEnvironment()
}

// CreateChildProcessEnv creates environment for child processes
func (em *EnvironmentManager) CreateChildProcessEnv() []string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var env []string
	for key, value := range em.environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// GetPath returns the PATH environment variable
func (em *EnvironmentManager) GetPath() string {
	return em.GetEnv("PATH")
}

// SetPath sets the PATH environment variable
func (em *EnvironmentManager) SetPath(path string) {
	em.SetEnv("PATH", path)
}

// AppendToPath appends a directory to PATH
func (em *EnvironmentManager) AppendToPath(dir string) {
	currentPath := em.GetPath()
	if currentPath == "" {
		em.SetPath(dir)
	} else {
		em.SetPath(fmt.Sprintf("%s:%s", currentPath, dir))
	}
}

// PrependToPath prepends a directory to PATH
func (em *EnvironmentManager) PrependToPath(dir string) {
	currentPath := em.GetPath()
	if currentPath == "" {
		em.SetPath(dir)
	} else {
		em.SetPath(fmt.Sprintf("%s:%s", dir, currentPath))
	}
}

// GetHomeDir returns the user's home directory
func (em *EnvironmentManager) GetHomeDir() string {
	home := em.GetEnv("HOME")
	if home == "" {
		home = em.originalEnv["HOME"]
	}
	return home
}

// GetUsername returns the current username
func (em *EnvironmentManager) GetUsername() string {
	user := em.GetEnv("USER")
	if user == "" {
		user = em.originalEnv["USER"]
	}
	return user
}

// CreateSession creates a new session environment
func (em *EnvironmentManager) CreateSession() *Session {
	em.mu.Lock()
	defer em.mu.Unlock()

	session := &Session{
		workingDir:  em.workingDir,
		environment: make(map[string]string),
	}

	// Copy current environment
	for k, v := range em.environment {
		session.environment[k] = v
	}

	return session
}

// Session represents a isolated environment session
type Session struct {
	workingDir  string
	environment map[string]string
}

// GetWorkingDir returns the session's working directory
func (s *Session) GetWorkingDir() string {
	return s.workingDir
}

// GetEnv gets an environment variable from the session
func (s *Session) GetEnv(key string) string {
	return s.environment[key]
}

// SetEnv sets an environment variable in the session
func (s *Session) SetEnv(key, value string) {
	s.environment[key] = value
}

// ApplySession applies the session environment to the manager
func (em *EnvironmentManager) ApplySession(session *Session) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.workingDir = session.workingDir
	em.environment = session.environment
}
