package main

import (
	"fmt"
	"log"

	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

func main() {
	// Create a new standard library instance
	stdlib := stdlib.New()

	fmt.Println("Testing Shode Standard Library")
	fmt.Println("==============================")

	// Test file system functions
	fmt.Println("\n1. File System Functions:")
	
	// Test working directory
	wd, err := stdlib.WorkingDir()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
	} else {
		fmt.Printf("Working directory: %s\n", wd)
	}

	// Test listing files
	files, err := stdlib.ListFiles(".")
	if err != nil {
		log.Printf("Error listing files: %v", err)
	} else {
		fmt.Printf("Files in current directory: %v\n", files)
	}

	// Test file existence
	goModExists := stdlib.FileExists("go.mod")
	fmt.Printf("go.mod exists: %t\n", goModExists)

	// Test string functions
	fmt.Println("\n2. String Functions:")
	
	testString := "  Hello, Shode!  "
	fmt.Printf("Original: '%s'\n", testString)
	fmt.Printf("Trimmed: '%s'\n", stdlib.Trim(testString))
	fmt.Printf("Uppercase: '%s'\n", stdlib.ToUpper(testString))
	fmt.Printf("Lowercase: '%s'\n", stdlib.ToLower(testString))
	
	containsHello := stdlib.Contains(testString, "Hello")
	fmt.Printf("Contains 'Hello': %t\n", containsHello)

	replaced := stdlib.Replace(testString, "Shode", "World")
	fmt.Printf("Replaced: '%s'\n", replaced)

	// Test environment functions
	fmt.Println("\n3. Environment Functions:")
	
	path := stdlib.GetEnv("PATH")
	if len(path) > 50 {
		path = path[:50] + "..."
	}
	fmt.Printf("PATH env var: %s\n", path)

	// Test utility functions
	fmt.Println("\n4. Utility Functions:")
	
	stdlib.Println("This is a test message printed via stdlib.Println")
	stdlib.Error("This is an error message printed via stdlib.Error\n")

	fmt.Println("Standard library test completed successfully!")
}
