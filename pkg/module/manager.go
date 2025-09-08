package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// ModuleManager manages Shode module loading and resolution
type ModuleManager struct {
	envManager *environment.EnvironmentManager
	parser     *parser.SimpleParser
	modules    map[string]*Module
}

// Module represents a loaded Shode module
type Module struct {
	Name     string
	Path     string
	Exports  map[string]*types.CommandNode
	Imports  map[string]*Module
	IsLoaded bool
}

// ModuleInfo contains information about a module
type ModuleInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description,omitempty"`
	Main        string            `json:"main,omitempty"`
	Exports     map[string]string `json:"exports,omitempty"`
}

// NewModuleManager creates a new module manager
func NewModuleManager() *ModuleManager {
	return &ModuleManager{
		envManager: environment.NewEnvironmentManager(),
		parser:     parser.NewSimpleParser(),
		modules:    make(map[string]*Module),
	}
}

// LoadModule loads a module from the given path
func (mm *ModuleManager) LoadModule(path string) (*Module, error) {
	// Check if module is already loaded
	if module, exists := mm.modules[path]; exists && module.IsLoaded {
		return module, nil
	}

	// Resolve absolute path
	absPath, err := mm.resolveModulePath(path)
	if err != nil {
		return nil, err
	}

	// Check if module exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("module not found: %s", path)
	}

	// Create new module
	module := &Module{
		Name:     filepath.Base(absPath),
		Path:     absPath,
		Exports:  make(map[string]*types.CommandNode),
		Imports:  make(map[string]*Module),
		IsLoaded: false,
	}

	// Load module exports
	if err := mm.loadModuleExports(module); err != nil {
		return nil, err
	}

	// Mark as loaded and store
	module.IsLoaded = true
	mm.modules[path] = module

	return module, nil
}

// resolveModulePath resolves a module path to an absolute path
func (mm *ModuleManager) resolveModulePath(path string) (string, error) {
	// Handle relative paths
	if !filepath.IsAbs(path) {
		wd := mm.envManager.GetWorkingDir()
		
		// Check if it's a local file
		localPath := filepath.Join(wd, path)
		if _, err := os.Stat(localPath); err == nil {
			return localPath, nil
		}

		// Check sh_models
		shModelsPath := filepath.Join(wd, "sh_models", path)
		if _, err := os.Stat(shModelsPath); err == nil {
			return shModelsPath, nil
		}

		return "", fmt.Errorf("module not found: %s", path)
	}

	return path, nil
}

// loadModuleExports loads exports from a module
func (mm *ModuleManager) loadModuleExports(module *Module) error {
	// Check for package.json first
	packageJsonPath := filepath.Join(module.Path, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		// TODO: Load package.json and handle main entry point
	}

	// Look for index.sh
	indexPath := filepath.Join(module.Path, "index.sh")
	if _, err := os.Stat(indexPath); err == nil {
		return mm.loadScriptExports(module, indexPath)
	}

	// Look for <module-name>.sh
	moduleScriptPath := filepath.Join(module.Path, module.Name+".sh")
	if _, err := os.Stat(moduleScriptPath); err == nil {
		return mm.loadScriptExports(module, moduleScriptPath)
	}

	return fmt.Errorf("no module entry point found in %s", module.Path)
}

// loadScriptExports loads exports from a script file
func (mm *ModuleManager) loadScriptExports(module *Module, scriptPath string) error {
	// Read script content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read module script: %v", err)
	}

	// Parse script
	script, err := mm.parser.ParseString(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse module script: %v", err)
	}

	// Extract exports (functions starting with export_)
	for _, node := range script.Nodes {
		if cmdNode, ok := node.(*types.CommandNode); ok {
			if strings.HasPrefix(cmdNode.Name, "export_") {
				exportName := strings.TrimPrefix(cmdNode.Name, "export_")
				module.Exports[exportName] = cmdNode
			}
		}
	}

	return nil
}

// Import imports a module and returns its exports
func (mm *ModuleManager) Import(path string) (map[string]*types.CommandNode, error) {
	module, err := mm.LoadModule(path)
	if err != nil {
		return nil, err
	}

	return module.Exports, nil
}

// GetModule returns a loaded module by path
func (mm *ModuleManager) GetModule(path string) (*Module, error) {
	module, exists := mm.modules[path]
	if !exists || !module.IsLoaded {
		return nil, fmt.Errorf("module not loaded: %s", path)
	}
	return module, nil
}

// ListModules returns all loaded modules
func (mm *ModuleManager) ListModules() []*Module {
	var modules []*Module
	for _, module := range mm.modules {
		if module.IsLoaded {
			modules = append(modules, module)
		}
	}
	return modules
}

// UnloadModule unloads a module
func (mm *ModuleManager) UnloadModule(path string) error {
	if _, exists := mm.modules[path]; !exists {
		return fmt.Errorf("module not found: %s", path)
	}
	delete(mm.modules, path)
	return nil
}

// ClearModules unloads all modules
func (mm *ModuleManager) ClearModules() {
	mm.modules = make(map[string]*Module)
}

// ResolveImport resolves an import statement
func (mm *ModuleManager) ResolveImport(importPath string) (string, error) {
	return mm.resolveModulePath(importPath)
}

// GetExport gets a specific export from a module
func (mm *ModuleManager) GetExport(modulePath, exportName string) (*types.CommandNode, error) {
	module, err := mm.GetModule(modulePath)
	if err != nil {
		return nil, err
	}

	// Try exact match first
	export, exists := module.Exports[exportName]
	if exists {
		return export, nil
	}

	// Try with parentheses for function-style exports
	export, exists = module.Exports[exportName+"()"]
	if exists {
		return export, nil
	}

	return nil, fmt.Errorf("export %s not found in module %s", exportName, modulePath)
}

// HasExport checks if a module has a specific export
func (mm *ModuleManager) HasExport(modulePath, exportName string) (bool, error) {
	module, err := mm.GetModule(modulePath)
	if err != nil {
		return false, err
	}

	// Try exact match first
	_, exists := module.Exports[exportName]
	if exists {
		return true, nil
	}

	// Try with parentheses for function-style exports
	_, exists = module.Exports[exportName+"()"]
	return exists, nil
}

// GetModuleInfo gets information about a module
func (mm *ModuleManager) GetModuleInfo(path string) (*ModuleInfo, error) {
	module, err := mm.GetModule(path)
	if err != nil {
		return nil, err
	}

	info := &ModuleInfo{
		Name:    module.Name,
		Exports: make(map[string]string),
	}

	// Collect export names
	for exportName := range module.Exports {
		info.Exports[exportName] = "function"
	}

	return info, nil
}
