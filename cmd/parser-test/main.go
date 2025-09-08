package main

import (
	"fmt"
	"log"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func main() {
	// Create a new simple parser
	p := parser.NewSimpleParser()

	// Test with a simple shell command
	testScript := `echo "Hello, World!"
ls -la
# This is a comment
cat /etc/passwd | grep root
echo 'quoted argument with spaces'`

	fmt.Println("Testing simple parser with script:")
	fmt.Println(testScript)
	fmt.Println("\n" + strings.Repeat("=", 50))

	// Parse the script
	script, err := p.ParseString(testScript)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	fmt.Printf("Parsed successfully! Found %d commands:\n", len(script.Nodes))
	for i, node := range script.Nodes {
		if cmd, ok := node.(*types.CommandNode); ok {
			fmt.Printf("%d: %s %v (line %d)\n", i+1, cmd.Name, cmd.Args, cmd.Pos.Line)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Parser debug output:")
	p.DebugPrint(testScript)
}
