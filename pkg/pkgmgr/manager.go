package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitee.com/com_818cloud/shode/pkg/environment"
)

// PackageManager manages Shode package dependencies
type PackageManager struct {
	envManager *environment.EnvironmentManager
	config     *PackageConfig
	configPath string
}

// PackageConfig represents the shode.json configuration
type PackageConfig struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Scripts      map[string]string `json:"scripts,omitempty"`
}

// PackageInfo represents information about an installed package
type PackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Main        string `json:"main,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	Repository  string `json:"repository,omitempty"`
}

// NewPackageManager creates a new package manager
func NewPackageManager() *PackageManager {
	return &PackageManager{
		envManager: environment.NewEnvironmentManager(),
		config:     &PackageConfig{},
	}
}

// Init initializes a new package configuration
func (pm *PackageManager) Init(name, version string) error {
	pm.config = &PackageConfig{
		Name:        name,
		Version:     version,
		Dependencies: make(map[string]string),
		DevDependencies: make(map[string]string),
		Scripts:     make(map[string]string),
	}

	// Set default config path
	wd := pm.envManager.GetWorkingDir()
	pm.configPath = filepath.Join(wd, "shode.json")

	return pm.SaveConfig()
}

// LoadConfig loads the package configuration from shode.json
func (pm *PackageManager) LoadConfig() error {
	wd := pm.envManager.GetWorkingDir()
	configPath := filepath.Join(wd, "shode.json")
	pm.configPath = configPath

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("shode.json not found. Run 'shode pkg init' first")
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read shode.json: %v", err)
	}

	if err := json.Unmarshal(data, &pm.config); err != nil {
		return fmt.Errorf("failed to parse shode.json: %v", err)
	}

	// Initialize maps if they are nil
	if pm.config.Dependencies == nil {
		pm.config.Dependencies = make(map[string]string)
	}
	if pm.config.DevDependencies == nil {
		pm.config.DevDependencies = make(map[string]string)
	}
	if pm.config.Scripts == nil {
		pm.config.Scripts = make(map[string]string)
	}

	return nil
}

// SaveConfig saves the package configuration to shode.json
func (pm *PackageManager) SaveConfig() error {
	if pm.configPath == "" {
		return fmt.Errorf("config path not set")
	}

	data, err := json.MarshalIndent(pm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	return ioutil.WriteFile(pm.configPath, data, 0644)
}

// AddDependency adds a package dependency
func (pm *PackageManager) AddDependency(name, version string, dev bool) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	if dev {
		pm.config.DevDependencies[name] = version
	} else {
		pm.config.Dependencies[name] = version
	}

	return pm.SaveConfig()
}

// RemoveDependency removes a package dependency
func (pm *PackageManager) RemoveDependency(name string, dev bool) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	if dev {
		delete(pm.config.DevDependencies, name)
	} else {
		delete(pm.config.Dependencies, name)
	}

	return pm.SaveConfig()
}

// AddScript adds a script to the configuration
func (pm *PackageManager) AddScript(name, command string) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	pm.config.Scripts[name] = command
	return pm.SaveConfig()
}

// RemoveScript removes a script from the configuration
func (pm *PackageManager) RemoveScript(name string) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	delete(pm.config.Scripts, name)
	return pm.SaveConfig()
}

// Install installs all dependencies
func (pm *PackageManager) Install() error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	fmt.Println("Installing dependencies...")

	// Create sh_models directory if it doesn't exist
	wd := pm.envManager.GetWorkingDir()
	shModelsPath := filepath.Join(wd, "sh_models")
	if err := os.MkdirAll(shModelsPath, 0755); err != nil {
		return fmt.Errorf("failed to create sh_models directory: %v", err)
	}

	// Install dependencies
	allDeps := make(map[string]string)
	for name, version := range pm.config.Dependencies {
		allDeps[name] = version
	}
	for name, version := range pm.config.DevDependencies {
		allDeps[name] = version
	}

	for name, version := range allDeps {
		fmt.Printf("Installing %s@%s\n", name, version)
		if err := pm.installPackage(name, version); err != nil {
			return fmt.Errorf("failed to install %s: %v", name, err)
		}
	}

	fmt.Println("All dependencies installed successfully!")
	return nil
}

// installPackage installs a single package
func (pm *PackageManager) installPackage(name, version string) error {
	wd := pm.envManager.GetWorkingDir()

	// For now, we'll simulate package installation
	// In a real implementation, this would download from a registry
	packagePath := filepath.Join(wd, "sh_models", name)
	if err := os.MkdirAll(packagePath, 0755); err != nil {
		return err
	}

	// Create a simple package.json for the installed package
	packageInfo := PackageInfo{
		Name:    name,
		Version: version,
		Main:    "index.sh",
	}

	infoData, err := json.MarshalIndent(packageInfo, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(packagePath, "package.json"), infoData, 0644); err != nil {
		return err
	}

	// Create a simple index.sh file
	indexContent := fmt.Sprintf(`#!/bin/sh
# %s v%s - Shode package
echo "Package %s version %s is installed"
`, name, version, name, version)

	if err := ioutil.WriteFile(filepath.Join(packagePath, "index.sh"), []byte(indexContent), 0755); err != nil {
		return err
	}

	return nil
}

// RunScript runs a script from the configuration
func (pm *PackageManager) RunScript(name string) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	script, exists := pm.config.Scripts[name]
	if !exists {
		return fmt.Errorf("script '%s' not found in shode.json", name)
	}

	fmt.Printf("Running script: %s\n", script)
	fmt.Println("(Script execution will be implemented in the execution engine)")

	return nil
}

// ListDependencies lists all dependencies
func (pm *PackageManager) ListDependencies() error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	fmt.Println("Dependencies:")
	for name, version := range pm.config.Dependencies {
		fmt.Printf("  %s: %s\n", name, version)
	}

	fmt.Println("\nDev Dependencies:")
	for name, version := range pm.config.DevDependencies {
		fmt.Printf("  %s: %s\n", name, version)
	}

	return nil
}

// GetConfig returns the current package configuration
func (pm *PackageManager) GetConfig() *PackageConfig {
	return pm.config
}

// GetConfigPath returns the path to the config file
func (pm *PackageManager) GetConfigPath() string {
	return pm.configPath
}
