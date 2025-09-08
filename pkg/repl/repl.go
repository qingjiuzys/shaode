package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// REPL represents a Read-Eval-Print Loop interactive environment
type REPL struct {
	envManager   *environment.EnvironmentManager
	security     *sandbox.SecurityChecker
	parser       *parser.SimpleParser
	stdlib       *stdlib.StdLib
	history      []string
	running      bool
}

// NewREPL creates a new interactive REPL environment
func NewREPL() *REPL {
	return &REPL{
		envManager: environment.NewEnvironmentManager(),
		security:   sandbox.NewSecurityChecker(),
		parser:     parser.NewSimpleParser(),
		stdlib:     stdlib.New(),
		history:    make([]string, 0),
		running:    false,
	}
}

// Start begins the REPL interactive session
func (r *REPL) Start() {
	r.running = true
	fmt.Println("Shode REPL - Interactive Shell Environment")
	fmt.Println("Type 'exit' or 'quit' to exit, 'help' for help")
	fmt.Printf("Working directory: %s\n", r.envManager.GetWorkingDir())

	scanner := bufio.NewScanner(os.Stdin)

	for r.running {
		fmt.Printf("shode> ")
		
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Add to history
		r.history = append(r.history, input)

		// Handle special commands
		if r.handleSpecialCommand(input) {
			continue
		}

		// Process the command
		r.processCommand(input)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

// handleSpecialCommand processes REPL-specific commands
func (r *REPL) handleSpecialCommand(input string) bool {
	switch strings.ToLower(input) {
	case "exit", "quit":
		r.running = false
		fmt.Println("Goodbye!")
		return true
	case "help":
		r.showHelp()
		return true
	case "clear":
		fmt.Print("\033[H\033[2J") // Clear screen
		return true
	case "pwd":
		fmt.Println(r.envManager.GetWorkingDir())
		return true
	case "env":
		r.showEnvironment()
		return true
	case "history":
		r.showHistory()
		return true
	case "cd":
		// Change to home directory if no argument
		home := r.envManager.GetHomeDir()
		if home != "" {
			r.envManager.ChangeDir(home)
			fmt.Println(home)
		}
		return true
	}

	// Handle cd with argument
	if strings.HasPrefix(input, "cd ") {
		dir := strings.TrimSpace(input[3:])
		if dir == "" {
			dir = r.envManager.GetHomeDir()
		}
		if err := r.envManager.ChangeDir(dir); err != nil {
			fmt.Printf("cd: %v\n", err)
		} else {
			fmt.Println(r.envManager.GetWorkingDir())
		}
		return true
	}

	return false
}

// processCommand processes a shell command
func (r *REPL) processCommand(input string) {
	// Parse the command
	script, err := r.parser.ParseString(input)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	if len(script.Nodes) == 0 {
		return
	}

	cmd := script.Nodes[0].(*types.CommandNode)

	// Check security
	if err := r.security.CheckCommand(cmd); err != nil {
		fmt.Printf("Security error: %v\n", err)
		return
	}

	// Execute the command
	r.executeCommand(cmd)
}

// executeCommand executes a parsed command
func (r *REPL) executeCommand(cmd *types.CommandNode) {
	commandName := strings.ToLower(cmd.Name)

	// Handle built-in commands
	switch commandName {
	case "echo":
		if len(cmd.Args) > 0 {
			fmt.Println(strings.Join(cmd.Args, " "))
		}
		return
	case "ls":
		r.handleLsCommand(cmd.Args)
		return
	case "cat":
		if len(cmd.Args) > 0 {
			r.handleCatCommand(cmd.Args[0])
		}
		return
	}

	// For other commands, just show what would be executed
	fmt.Printf("Would execute: %s %s\n", cmd.Name, strings.Join(cmd.Args, " "))
	fmt.Println("(Execution engine will handle this in future versions)")
}

// handleLsCommand handles the ls command
func (r *REPL) handleLsCommand(args []string) {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	files, err := r.stdlib.ListFiles(dir)
	if err != nil {
		fmt.Printf("ls: %v\n", err)
		return
	}

	for _, file := range files {
		fmt.Println(file)
	}
}

// handleCatCommand handles the cat command
func (r *REPL) handleCatCommand(filename string) {
	content, err := r.stdlib.ReadFile(filename)
	if err != nil {
		fmt.Printf("cat: %v\n", err)
		return
	}
	fmt.Print(content)
}

// showHelp displays REPL help information
func (r *REPL) showHelp() {
	fmt.Println("Shode REPL Commands:")
	fmt.Println("  help          - Show this help")
	fmt.Println("  exit, quit    - Exit the REPL")
	fmt.Println("  clear         - Clear the screen")
	fmt.Println("  pwd           - Show current directory")
	fmt.Println("  env           - Show environment variables")
	fmt.Println("  history       - Show command history")
	fmt.Println("  cd [dir]      - Change directory")
	fmt.Println("  ls [dir]      - List files")
	fmt.Println("  cat <file>    - Show file content")
	fmt.Println("  echo <text>   - Echo text")
	fmt.Println("  Other shell commands will be processed by Shode")
}

// showEnvironment displays current environment variables
func (r *REPL) showEnvironment() {
	env := r.envManager.GetAllEnv()
	for key, value := range env {
		fmt.Printf("%s=%s\n", key, value)
	}
}

// showHistory displays command history
func (r *REPL) showHistory() {
	for i, cmd := range r.history {
		fmt.Printf("%4d  %s\n", i+1, cmd)
	}
}

// Stop stops the REPL session
func (r *REPL) Stop() {
	r.running = false
}

// GetHistory returns the command history
func (r *REPL) GetHistory() []string {
	return r.history
}
