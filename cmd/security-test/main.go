package main

import (
	"fmt"
	"log"

	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func main() {
	// Create a new security checker
	security := sandbox.NewSecurityChecker()
	parser := parser.NewSimpleParser()

	fmt.Println("Testing Shode Security Checker")
	fmt.Println("==============================")

	// Test commands
	testCommands := []string{
		"echo hello world",                    // Safe command
		"rm -rf /",                            // Dangerous command
		"ls -la /etc/passwd",                  // Sensitive file access
		"cat /root/.bashrc",                   // Root directory access
		"iptables -L",                         // Network command
		"echo 'safe command'",                 // Safe with quotes
		"useradd testuser",                    // User management
		"passwd --password secret",            // Password in command line
	}

	for i, cmdText := range testCommands {
		fmt.Printf("\n%d. Testing command: %s\n", i+1, cmdText)

		// Parse the command
		script, err := parser.ParseString(cmdText)
		if err != nil {
			log.Printf("Error parsing command: %v", err)
			continue
		}

		if len(script.Nodes) == 0 {
			fmt.Println("  No commands parsed")
			continue
		}

		cmd := script.Nodes[0].(*types.CommandNode)

		// Generate security report
		report := security.GetSecurityReport(cmd)
		fmt.Printf("  Security report:\n")
		fmt.Printf("    Command: %s\n", report["command"])
		fmt.Printf("    Arguments: %v\n", report["arguments"])
		fmt.Printf("    Dangerous: %t\n", report["is_dangerous_command"])
		fmt.Printf("    Network: %t\n", report["is_network_command"])
		fmt.Printf("    Sensitive files: %v\n", report["sensitive_files"])

		// Check security
		err = security.CheckCommand(cmd)
		if err != nil {
			fmt.Printf("  ❌ SECURITY VIOLATION: %s\n", err)
		} else {
			fmt.Printf("  ✅ Security check passed\n")
		}
	}

	// Test custom security rules
	fmt.Println("\nTesting Custom Security Rules:")
	fmt.Println("------------------------------")

	// Add a custom dangerous command
	security.AddDangerousCommand("custom-dangerous")
	testCustom, _ := parser.ParseString("custom-dangerous --force")
	customCmd := testCustom.Nodes[0].(*types.CommandNode)

	err := security.CheckCommand(customCmd)
	if err != nil {
		fmt.Printf("Custom dangerous command blocked: %s\n", err)
	}

	// Remove a dangerous command and test
	security.RemoveDangerousCommand("rm")
	testRm, _ := parser.ParseString("rm temporary-file.txt")
	rmCmd := testRm.Nodes[0].(*types.CommandNode)

	err = security.CheckCommand(rmCmd)
	if err != nil {
		fmt.Printf("RM command still blocked: %s\n", err)
	} else {
		fmt.Printf("RM command allowed after removal from blacklist\n")
	}

	fmt.Println("\nSecurity testing completed!")
}
