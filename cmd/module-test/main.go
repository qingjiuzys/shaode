package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gitee.com/com_818cloud/shode/pkg/module"
)

func main() {
	fmt.Println("Testing Shode Module System")
	fmt.Println("===========================")

	// Create test module directory
	testDir := "./test-module"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("Failed to create test directory: %v\n", err)
		return
	}
	defer os.RemoveAll(testDir)

	// Create a test module script
	moduleScript := `#!/bin/sh
# Test module with exports

export_hello() {
    echo "Hello from module!"
}

export_greet() {
    echo "Greetings, $1!"
}

export_calculate() {
    echo $(($1 + $2))
}

# Non-exported function (should not be accessible)
internal_function() {
    echo "This is internal"
}
`

	scriptPath := filepath.Join(testDir, "index.sh")
	if err := os.WriteFile(scriptPath, []byte(moduleScript), 0755); err != nil {
		fmt.Printf("Failed to create module script: %v\n", err)
		return
	}

	// Test module manager
	mm := module.NewModuleManager()

	fmt.Println("1. Loading module...")
	testModule, err := mm.LoadModule(testDir)
	if err != nil {
		fmt.Printf("Failed to load module: %v\n", err)
		return
	}

	fmt.Printf("Module loaded: %s\n", testModule.Name)
	fmt.Printf("Module path: %s\n", testModule.Path)

	fmt.Println("\n2. Listing exports:")
	for exportName := range testModule.Exports {
		fmt.Printf("  - %s\n", exportName)
	}

	fmt.Println("\n3. Getting module info:")
	info, err := mm.GetModuleInfo(testDir)
	if err != nil {
		fmt.Printf("Failed to get module info: %v\n", err)
		return
	}

	fmt.Printf("Module name: %s\n", info.Name)
	fmt.Println("Exports:")
	for exportName, exportType := range info.Exports {
		fmt.Printf("  - %s: %s\n", exportName, exportType)
	}

	fmt.Println("\n4. Testing import function:")
	exports, err := mm.Import(testDir)
	if err != nil {
		fmt.Printf("Failed to import module: %v\n", err)
		return
	}

	fmt.Printf("Imported %d exports:\n", len(exports))
	for name := range exports {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\n5. Testing specific export retrieval:")
	// Debug: print all available exports with full details
	fmt.Println("Available exports with details:")
	for exportName, exportNode := range testModule.Exports {
		fmt.Printf("  - %s: %s (args: %v)\n", exportName, exportNode.Name, exportNode.Args)
	}

	// Debug: check what's actually in the exports map
	fmt.Println("\nDebug: Direct access to exports map:")
	for key, value := range testModule.Exports {
		fmt.Printf("  Key: '%s', Value: %s (args: %v)\n", key, value.Name, value.Args)
	}

	// Try to access exports directly from the module using the correct keys
	if helloCmd, exists := testModule.Exports["hello()"]; exists {
		fmt.Printf("Direct access - Hello export: %s (args: %v)\n", helloCmd.Name, helloCmd.Args)
	} else {
		fmt.Println("Direct access - hello() export not found")
	}

	if greetCmd, exists := testModule.Exports["greet()"]; exists {
		fmt.Printf("Direct access - Greet export: %s (args: %v)\n", greetCmd.Name, greetCmd.Args)
	} else {
		fmt.Println("Direct access - greet() export not found")
	}

	fmt.Println("\n6. Testing export existence check:")
	hasHello, err := mm.HasExport(testDir, "hello")
	if err != nil {
		fmt.Printf("Failed to check hello export: %v\n", err)
		return
	}
	fmt.Printf("Has hello export: %v\n", hasHello)

	hasInternal, err := mm.HasExport(testDir, "internal_function")
	if err != nil {
		fmt.Printf("Failed to check internal_function export: %v\n", err)
		return
	}
	fmt.Printf("Has internal_function export: %v (should be false)\n", hasInternal)

	fmt.Println("\n7. Testing module listing:")
	modules := mm.ListModules()
	fmt.Printf("Loaded modules: %d\n", len(modules))
	for _, mod := range modules {
		fmt.Printf("  - %s (%s)\n", mod.Name, mod.Path)
	}

	fmt.Println("\n8. Testing module unloading:")
	if err := mm.UnloadModule(testDir); err != nil {
		fmt.Printf("Failed to unload module: %v\n", err)
		return
	}
	fmt.Println("Module unloaded successfully")

	// Try to get unloaded module
	_, err = mm.GetModule(testDir)
	if err != nil {
		fmt.Printf("Expected error getting unloaded module: %v\n", err)
	}

	fmt.Println("\nModule system testing completed successfully!")
}
