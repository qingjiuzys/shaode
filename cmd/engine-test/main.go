package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

func main() {
	fmt.Println("Testing Shode Execution Engine")
	fmt.Println("==============================")

	// Initialize all required components
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()
	parser := parser.NewSimpleParser()

	// Create execution engine
	execEngine := engine.NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	// Test cases
	testCases := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "Simple echo command",
			script:   `echo "Hello, World!"`,
			expected: "Hello, World!",
		},
		{
			name:     "Standard library function",
			script:   `Print "Hello from stdlib"`,
			expected: "Hello from stdlib",
		},
		{
			name:     "Multiple commands",
			script:   `echo "First"; echo "Second"`,
			expected: "First\nSecond",
		},
		{
			name:     "Environment variable",
			script:   `echo "Home: $HOME"`,
			expected: "Home: " + envManager.GetEnv("HOME"),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		fmt.Printf("\nTesting: %s\n", tc.name)
		fmt.Printf("Script: %s\n", tc.script)

		// Parse script
		script, err := parser.ParseString(tc.script)
		if err != nil {
			log.Printf("Failed to parse script: %v", err)
			continue
		}

		// Execute script
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		startTime := time.Now()
		result, err := execEngine.Execute(ctx, script)
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("Execution failed: %v", err)
			continue
		}

		// Print results
		fmt.Printf("Success: %v\n", result.Success)
		fmt.Printf("Exit code: %d\n", result.ExitCode)
		fmt.Printf("Output: %q\n", result.Output)
		fmt.Printf("Duration: %v\n", duration)

		// Print individual command results
		for i, cmdResult := range result.Commands {
			fmt.Printf("  Command %d: %s\n", i+1, cmdResult.Command.Name)
			fmt.Printf("    Success: %v, Exit: %d, Mode: %v\n",
				cmdResult.Success, cmdResult.ExitCode, cmdResult.Mode)
			if cmdResult.Output != "" {
				fmt.Printf("    Output: %q\n", cmdResult.Output)
			}
			if cmdResult.Error != "" {
				fmt.Printf("    Error: %q\n", cmdResult.Error)
			}
		}

		// Check if expected output matches
		if result.Output == tc.expected {
			fmt.Printf("✅ PASS: Output matches expected\n")
		} else {
			fmt.Printf("❌ FAIL: Expected %q, got %q\n", tc.expected, result.Output)
		}
	}

	// Test security features
	fmt.Printf("\nTesting Security Features\n")
	fmt.Printf("========================\n")

	securityTests := []struct {
		name        string
		script      string
		shouldBlock bool
	}{
		{
			name:        "Dangerous command (rm)",
			script:      `rm -rf /`,
			shouldBlock: true,
		},
		{
			name:        "Safe command",
			script:      `echo "safe"`,
			shouldBlock: false,
		},
	}

	for _, test := range securityTests {
		fmt.Printf("\nTesting: %s\n", test.name)
		fmt.Printf("Script: %s\n", test.script)

		script, err := parser.ParseString(test.script)
		if err != nil {
			log.Printf("Failed to parse: %v", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := execEngine.Execute(ctx, script)

		if test.shouldBlock {
			if result != nil && !result.Success && result.Commands[0].Error != "" {
				fmt.Printf("✅ PASS: Command was blocked by security\n")
			} else {
				fmt.Printf("❌ FAIL: Command was not blocked\n")
			}
		} else {
			if result != nil && result.Success {
				fmt.Printf("✅ PASS: Command executed successfully\n")
			} else {
				fmt.Printf("❌ FAIL: Command should have executed\n")
			}
		}
	}

	// Test performance with cache
	fmt.Printf("\nTesting Performance with Cache\n")
	fmt.Printf("=============================\n")

	perfScript := `echo "performance test"`
	script, err := parser.ParseString(perfScript)
	if err != nil {
		log.Fatalf("Failed to parse performance script: %v", err)
	}

	// First execution (cold)
	ctx := context.Background()
	start1 := time.Now()
	result1, err := execEngine.Execute(ctx, script)
	duration1 := time.Since(start1)

	if err != nil {
		log.Fatalf("First execution failed: %v", err)
	}

	// Second execution (should be cached)
	start2 := time.Now()
	result2, err := execEngine.Execute(ctx, script)
	duration2 := time.Since(start2)

	if err != nil {
		log.Fatalf("Second execution failed: %v", err)
	}

	fmt.Printf("First execution: %v\n", duration1)
	fmt.Printf("Second execution: %v\n", duration2)
	fmt.Printf("Speedup: %.2fx\n", float64(duration1)/float64(duration2))

	if result1.Success && result2.Success {
		fmt.Printf("✅ PASS: Both executions successful\n")
	} else {
		fmt.Printf("❌ FAIL: Executions failed\n")
	}

	fmt.Printf("\nExecution engine testing completed!\n")
}
