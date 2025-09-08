package main

import (
	"fmt"

	"gitee.com/com_818cloud/shode/pkg/environment"
)

func main() {
	// Create a new environment manager
	env := environment.NewEnvironmentManager()

	fmt.Println("Testing Shode Environment Manager")
	fmt.Println("=================================")

	// Test initial state
	fmt.Printf("Initial working directory: %s\n", env.GetWorkingDir())
	fmt.Printf("Initial USER: %s\n", env.GetEnv("USER"))
	fmt.Printf("Initial HOME: %s\n", env.GetHomeDir())
	fmt.Printf("Initial PATH: %s\n", env.GetPath())

	// Test environment variable manipulation
	fmt.Println("\n1. Environment Variable Tests:")
	
	// Set a new environment variable
	env.SetEnv("SHODE_TEST", "test_value")
	fmt.Printf("Set SHODE_TEST: %s\n", env.GetEnv("SHODE_TEST"))

	// Test getting all environment variables
	allEnv := env.GetAllEnv()
	fmt.Printf("Total environment variables: %d\n", len(allEnv))

	// Test unsetting environment variable
	env.UnsetEnv("SHODE_TEST")
	fmt.Printf("After unset SHODE_TEST: %s\n", env.GetEnv("SHODE_TEST"))

	// Test PATH manipulation
	fmt.Println("\n2. PATH Manipulation Tests:")
	
	originalPath := env.GetPath()
	fmt.Printf("Original PATH: %s\n", originalPath)

	// Append to PATH
	env.AppendToPath("/custom/bin")
	fmt.Printf("After appending /custom/bin: %s\n", env.GetPath())

	// Prepend to PATH
	env.PrependToPath("/usr/local/custom/bin")
	fmt.Printf("After prepending /usr/local/custom/bin: %s\n", env.GetPath())

	// Reset PATH
	env.SetPath(originalPath)
	fmt.Printf("After resetting PATH: %s\n", env.GetPath())

	// Test directory changing
	fmt.Println("\n3. Directory Change Tests:")
	
	// Try to change to a non-existent directory
	err := env.ChangeDir("/non/existent/directory")
	if err != nil {
		fmt.Printf("Error changing to non-existent directory: %v\n", err)
	}

	// Change to a valid directory (current directory should work)
	err = env.ChangeDir(".")
	if err != nil {
		fmt.Printf("Error changing to current directory: %v\n", err)
	} else {
		fmt.Printf("Successfully changed to current directory: %s\n", env.GetWorkingDir())
	}

	// Test session management
	fmt.Println("\n4. Session Management Tests:")
	
	// Create a session
	session := env.CreateSession()
	fmt.Printf("Session working directory: %s\n", session.GetWorkingDir())

	// Modify session environment
	session.SetEnv("SESSION_VAR", "session_value")
	fmt.Printf("Session variable set: %s\n", session.GetEnv("SESSION_VAR"))

	// Session should not affect main environment yet
	fmt.Printf("Main environment before applying session: %s\n", env.GetEnv("SESSION_VAR"))

	// Apply session to main environment
	env.ApplySession(session)
	fmt.Printf("Main environment after applying session: %s\n", env.GetEnv("SESSION_VAR"))

	// Test child process environment
	fmt.Println("\n5. Child Process Environment Tests:")
	
	childEnv := env.CreateChildProcessEnv()
	fmt.Printf("Child environment variables count: %d\n", len(childEnv))
	
	// Show first few environment variables
	fmt.Println("First 5 child environment variables:")
	for i := 0; i < 5 && i < len(childEnv); i++ {
		fmt.Printf("  %s\n", childEnv[i])
	}

	// Test environment restoration
	fmt.Println("\n6. Environment Restoration Test:")
	
	// Modify environment
	env.SetEnv("TEMPORARY_VAR", "temp_value")
	fmt.Printf("Set temporary variable: %s\n", env.GetEnv("TEMPORARY_VAR"))

	// Restore original environment
	env.RestoreOriginalEnvironment()
	fmt.Printf("After restoration, TEMPORARY_VAR: %s\n", env.GetEnv("TEMPORARY_VAR"))
	fmt.Printf("Original USER preserved: %s\n", env.GetEnv("USER"))

	fmt.Println("\nEnvironment manager testing completed successfully!")
}
