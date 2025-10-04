package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// ExecutionMode represents the execution mode for commands
type ExecutionMode int

const (
	ModeInterpreted ExecutionMode = iota // Interpret built-in functions
	ModeProcess                          // Execute external processes
	ModeHybrid                           // Smart hybrid execution
)

// ExecutionEngine is the core engine for executing shell commands
type ExecutionEngine struct {
	envManager  *environment.EnvironmentManager
	stdlib      *stdlib.StdLib
	moduleMgr   *module.ModuleManager
	security    *sandbox.SecurityChecker
	processPool *ProcessPool
	cache       *CommandCache
}

// ExecutionResult represents the result of executing an AST
type ExecutionResult struct {
	Success    bool
	ExitCode   int
	Output     string
	Error      string
	Duration   time.Duration
	Commands   []*CommandResult
}

// CommandResult represents the result of a single command execution
type CommandResult struct {
	Command   *types.CommandNode
	Success   bool
	ExitCode  int
	Output    string
	Error     string
	Duration  time.Duration
	Mode      ExecutionMode
}

// PipelineResult represents the result of pipeline execution
type PipelineResult struct {
	Success  bool
	ExitCode int
	Output   string
	Error    string
	Results  []*CommandResult
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(
	envManager *environment.EnvironmentManager,
	stdlib *stdlib.StdLib,
	moduleMgr *module.ModuleManager,
	security *sandbox.SecurityChecker,
) *ExecutionEngine {
	return &ExecutionEngine{
		envManager: envManager,
		stdlib:     stdlib,
		moduleMgr:  moduleMgr,
		security:   security,
		processPool: NewProcessPool(10, 30*time.Second),
		cache:       NewCommandCache(1000),
	}
}

// Execute executes a complete script
func (ee *ExecutionEngine) Execute(ctx context.Context, script *types.ScriptNode) (*ExecutionResult, error) {
	startTime := time.Now()
	
	result := &ExecutionResult{
		Commands: make([]*CommandResult, 0, len(script.Nodes)),
	}

	for _, node := range script.Nodes {
		switch n := node.(type) {
		case *types.CommandNode:
			cmdResult, err := ee.ExecuteCommand(ctx, n)
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, cmdResult)
			
			if !cmdResult.Success {
				result.Success = false
				result.ExitCode = cmdResult.ExitCode
				break
			}

		case *types.PipeNode:
			// Execute pipeline
			pipeResult, err := ee.ExecutePipeline(ctx, n)
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, pipeResult.Results...)
			
			if !pipeResult.Success {
				result.Success = false
				result.ExitCode = pipeResult.ExitCode
				break
			}

		case *types.IfNode:
			// Execute if-then-else
			ifResult, err := ee.ExecuteIf(ctx, n)
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, ifResult.Commands...)
			
			if !ifResult.Success {
				result.Success = false
				result.ExitCode = ifResult.ExitCode
				break
			}

		case *types.ForNode:
			// Execute for loop
			forResult, err := ee.ExecuteFor(ctx, n)
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, forResult.Commands...)
			
			if !forResult.Success {
				result.Success = false
				result.ExitCode = forResult.ExitCode
				break
			}

		case *types.WhileNode:
			// Execute while loop
			whileResult, err := ee.ExecuteWhile(ctx, n)
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, whileResult.Commands...)
			
			if !whileResult.Success {
				result.Success = false
				result.ExitCode = whileResult.ExitCode
				break
			}

		case *types.AssignmentNode:
			// Execute variable assignment
			ee.envManager.SetEnv(n.Name, n.Value)

		case *types.FunctionNode:
			// Store function definition (not executing it)
			// TODO: Implement function storage and execution
			
		default:
			return nil, fmt.Errorf("unsupported node type: %T", n)
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = true
	return result, nil
}

// ExecuteCommand executes a single command
func (ee *ExecutionEngine) ExecuteCommand(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	startTime := time.Now()

	// Security check
	if err := ee.security.CheckCommand(cmd); err != nil {
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    fmt.Sprintf("Security violation: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// Decide execution mode
	mode := ee.decideExecutionMode(cmd)

	var result *CommandResult
	var err error

	switch mode {
	case ModeInterpreted:
		result, err = ee.executeInterpreted(ctx, cmd)
	case ModeProcess:
		result, err = ee.executeProcess(ctx, cmd)
	case ModeHybrid:
		result, err = ee.executeHybrid(ctx, cmd)
	default:
		return nil, fmt.Errorf("unknown execution mode: %v", mode)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime)
	result.Mode = mode
	return result, nil
}

// ExecutePipeline executes a pipeline of commands with proper data flow
func (ee *ExecutionEngine) ExecutePipeline(ctx context.Context, pipeline *types.PipeNode) (*PipelineResult, error) {
	// Collect all commands in the pipeline
	commands := ee.collectPipelineCommands(pipeline)
	results := make([]*CommandResult, 0, len(commands))
	
	// Execute commands with piped data flow
	var previousOutput string
	for i, cmd := range commands {
		var result *CommandResult
		var err error
		
		if i == 0 {
			// First command - execute normally
			result, err = ee.ExecuteCommand(ctx, cmd)
		} else {
			// Subsequent commands - use previous output as input
			result, err = ee.ExecuteCommandWithInput(ctx, cmd, previousOutput)
		}
		
		if err != nil {
			return nil, err
		}
		
		results = append(results, result)
		
		// If command failed, stop pipeline
		if !result.Success {
			return &PipelineResult{
				Success:  false,
				ExitCode: result.ExitCode,
				Output:   result.Output,
				Error:    result.Error,
				Results:  results,
			}, nil
		}
		
		// Store output for next command
		previousOutput = result.Output
	}
	
	// Return final result
	lastResult := results[len(results)-1]
	return &PipelineResult{
		Success:  true,
		ExitCode: 0,
		Output:   lastResult.Output,
		Error:    "",
		Results:  results,
	}, nil
}

// collectPipelineCommands collects all commands from a pipeline tree
func (ee *ExecutionEngine) collectPipelineCommands(node types.Node) []*types.CommandNode {
	var commands []*types.CommandNode
	
	switch n := node.(type) {
	case *types.PipeNode:
		// Recursively collect left commands
		commands = append(commands, ee.collectPipelineCommands(n.Left)...)
		// Recursively collect right commands
		commands = append(commands, ee.collectPipelineCommands(n.Right)...)
	case *types.CommandNode:
		commands = append(commands, n)
	}
	
	return commands
}

// ExecuteCommandWithInput executes a command with input data
func (ee *ExecutionEngine) ExecuteCommandWithInput(ctx context.Context, cmd *types.CommandNode, input string) (*CommandResult, error) {
	startTime := time.Now()
	
	// Security check
	if err := ee.security.CheckCommand(cmd); err != nil {
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    fmt.Sprintf("Security violation: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}
	
	// Execute process with input
	result, err := ee.executeProcessWithInput(ctx, cmd, input)
	if err != nil {
		return nil, err
	}
	
	result.Duration = time.Since(startTime)
	result.Mode = ModeProcess
	return result, nil
}

// executeProcessWithInput executes a command with stdin input
func (ee *ExecutionEngine) executeProcessWithInput(ctx context.Context, cmd *types.CommandNode, input string) (*CommandResult, error) {
	// Create command with context
	command := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	
	// Set environment
	envVars := make([]string, 0, len(ee.envManager.GetAllEnv()))
	for key, value := range ee.envManager.GetAllEnv() {
		envVars = append(envVars, key+"="+value)
	}
	command.Env = envVars
	command.Dir = ee.envManager.GetWorkingDir()
	
	// Set up pipes
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	
	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr
	
	// Start command
	if err := command.Start(); err != nil {
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    err.Error(),
		}, nil
	}
	
	// Write input to stdin
	if _, err := stdin.Write([]byte(input)); err != nil {
		return nil, fmt.Errorf("failed to write to stdin: %v", err)
	}
	stdin.Close()
	
	// Wait for command to complete
	err = command.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	
	return &CommandResult{
		Command:  cmd,
		Success:  err == nil,
		ExitCode: exitCode,
		Output:   stdout.String(),
		Error:    stderr.String(),
	}, nil
}

// decideExecutionMode determines the best execution mode for a command
func (ee *ExecutionEngine) decideExecutionMode(cmd *types.CommandNode) ExecutionMode {
	// Check if it's a standard library function
	if ee.isStdLibFunction(cmd.Name) {
		return ModeInterpreted
	}

	// Check if it's a module export (TODO: implement module export check)
	// if ee.moduleMgr.IsExportedFunction(cmd.Name) {
	//     return ModeInterpreted
	// }

	// Check if external command exists
	if ee.isExternalCommandAvailable(cmd.Name) {
		return ModeProcess
	}

	// Default to process execution
	return ModeProcess
}

// isStdLibFunction checks if a function exists in standard library
func (ee *ExecutionEngine) isStdLibFunction(funcName string) bool {
	// Map of standard library functions
	stdlibFunctions := map[string]bool{
		"Print":      true,
		"Println":    true,
		"Error":      true,
		"Errorln":    true,
		"ReadFile":   true,
		"WriteFile":  true,
		"ListFiles":  true,
		"FileExists": true,
		"Contains":   true,
		"Replace":    true,
		"ToUpper":    true,
		"ToLower":    true,
		"Trim":       true,
		"GetEnv":     true,
		"SetEnv":     true,
		"WorkingDir": true,
		"ChangeDir":  true,
	}
	return stdlibFunctions[funcName]
}

// executeInterpreted executes a command using the interpreter (built-in functions)
func (ee *ExecutionEngine) executeInterpreted(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Execute using standard library
	result, err := ee.executeStdLibFunction(cmd.Name, cmd.Args)
	if err != nil {
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    err.Error(),
		}, nil
	}

	return &CommandResult{
		Command:  cmd,
		Success:  true,
		ExitCode: 0,
		Output:   result,
	}, nil
}

// executeStdLibFunction executes a standard library function
func (ee *ExecutionEngine) executeStdLibFunction(funcName string, args []string) (string, error) {
	switch funcName {
	case "Print":
		if len(args) > 0 {
			ee.stdlib.Print(args[0])
			return args[0], nil
		}
		return "", nil
	case "Println":
		if len(args) > 0 {
			ee.stdlib.Println(args[0])
			return args[0], nil
		}
		ee.stdlib.Println("")
		return "", nil
	case "Error":
		if len(args) > 0 {
			ee.stdlib.Error(args[0])
			return args[0], nil
		}
		return "", nil
	case "Errorln":
		if len(args) > 0 {
			ee.stdlib.Errorln(args[0])
			return args[0], nil
		}
		ee.stdlib.Errorln("")
		return "", nil
	case "ReadFile":
		if len(args) == 0 {
			return "", fmt.Errorf("ReadFile requires filename argument")
		}
		return ee.stdlib.ReadFile(args[0])
	case "WriteFile":
		if len(args) < 2 {
			return "", fmt.Errorf("WriteFile requires filename and content arguments")
		}
		err := ee.stdlib.WriteFile(args[0], args[1])
		return "File written", err
	case "ListFiles":
		if len(args) == 0 {
			files, err := ee.stdlib.ListFiles(".")
			if err != nil {
				return "", err
			}
			return strings.Join(files, "\n"), nil
		}
		files, err := ee.stdlib.ListFiles(args[0])
		if err != nil {
			return "", err
		}
		return strings.Join(files, "\n"), nil
	case "FileExists":
		if len(args) == 0 {
			return "", fmt.Errorf("FileExists requires filename argument")
		}
		exists := ee.stdlib.FileExists(args[0])
		return fmt.Sprintf("%v", exists), nil
	case "Contains":
		if len(args) < 2 {
			return "", fmt.Errorf("Contains requires haystack and needle arguments")
		}
		contains := ee.stdlib.Contains(args[0], args[1])
		return fmt.Sprintf("%v", contains), nil
	case "Replace":
		if len(args) < 3 {
			return "", fmt.Errorf("Replace requires string, old, and new arguments")
		}
		return ee.stdlib.Replace(args[0], args[1], args[2]), nil
	case "ToUpper":
		if len(args) == 0 {
			return "", nil
		}
		return ee.stdlib.ToUpper(args[0]), nil
	case "ToLower":
		if len(args) == 0 {
			return "", nil
		}
		return ee.stdlib.ToLower(args[0]), nil
	case "Trim":
		if len(args) == 0 {
			return "", nil
		}
		return ee.stdlib.Trim(args[0]), nil
	case "GetEnv":
		if len(args) == 0 {
			return "", fmt.Errorf("GetEnv requires environment variable name")
		}
		return ee.stdlib.GetEnv(args[0]), nil
	case "SetEnv":
		if len(args) < 2 {
			return "", fmt.Errorf("SetEnv requires key and value arguments")
		}
		err := ee.stdlib.SetEnv(args[0], args[1])
		return "Environment variable set", err
	case "WorkingDir":
		wd, err := ee.stdlib.WorkingDir()
		if err != nil {
			return "", err
		}
		return wd, nil
	case "ChangeDir":
		if len(args) == 0 {
			return "", fmt.Errorf("ChangeDir requires directory path")
		}
		err := ee.stdlib.ChangeDir(args[0])
		return "Directory changed", err
	default:
		return "", fmt.Errorf("unknown standard library function: %s", funcName)
	}
}

// executeProcess executes a command as an external process
func (ee *ExecutionEngine) executeProcess(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Check cache first (only if no redirects)
	if cmd.Redirect == nil {
		if cached, ok := ee.cache.Get(cmd.Name, cmd.Args); ok {
			return cached, nil
		}
	}

	// Create command with context
	command := exec.CommandContext(ctx, cmd.Name, cmd.Args...)

	// Set environment - convert map[string]string to []string
	envVars := make([]string, 0, len(ee.envManager.GetAllEnv()))
	for key, value := range ee.envManager.GetAllEnv() {
		envVars = append(envVars, key+"="+value)
	}
	command.Env = envVars

	// Set working directory
	command.Dir = ee.envManager.GetWorkingDir()

	// Handle redirections
	var stdout, stderr strings.Builder
	if cmd.Redirect != nil {
		if err := ee.setupRedirect(command, cmd.Redirect, &stdout, &stderr); err != nil {
			return &CommandResult{
				Command:  cmd,
				Success:  false,
				ExitCode: 1,
				Error:    fmt.Sprintf("redirect error: %v", err),
			}, nil
		}
	} else {
		// No redirect - capture output
		command.Stdout = &stdout
		command.Stderr = &stderr
	}

	// Execute command
	startTime := time.Now()
	err := command.Run()
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	result := &CommandResult{
		Command:  cmd,
		Success:  err == nil,
		ExitCode: exitCode,
		Output:   stdout.String(),
		Error:    stderr.String(),
		Duration: duration,
	}

	// Cache successful results (only if no redirects)
	if err == nil && cmd.Redirect == nil {
		ee.cache.Put(cmd.Name, cmd.Args, result)
	}

	return result, nil
}

// setupRedirect sets up input/output redirection for a command
func (ee *ExecutionEngine) setupRedirect(cmd *exec.Cmd, redirect *types.RedirectNode, stdout, stderr *strings.Builder) error {
	switch redirect.Op {
	case ">": // Output redirection (overwrite)
		file, err := os.Create(redirect.File)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", redirect.File, err)
		}
		defer file.Close()
		
		if redirect.Fd == 1 || redirect.Fd == 0 { // stdout
			cmd.Stdout = file
		} else if redirect.Fd == 2 { // stderr
			cmd.Stderr = file
		}
		
	case ">>": // Output redirection (append)
		file, err := os.OpenFile(redirect.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", redirect.File, err)
		}
		defer file.Close()
		
		if redirect.Fd == 1 || redirect.Fd == 0 {
			cmd.Stdout = file
		} else if redirect.Fd == 2 {
			cmd.Stderr = file
		}
		
	case "<": // Input redirection
		file, err := os.Open(redirect.File)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", redirect.File, err)
		}
		defer file.Close()
		cmd.Stdin = file
		
	case "2>&1": // Redirect stderr to stdout
		cmd.Stderr = cmd.Stdout
		
	case "&>": // Redirect both stdout and stderr to file
		file, err := os.Create(redirect.File)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", redirect.File, err)
		}
		defer file.Close()
		cmd.Stdout = file
		cmd.Stderr = file
		
	default:
		return fmt.Errorf("unsupported redirect operator: %s", redirect.Op)
	}
	
	return nil
}

// executeHybrid executes a command using hybrid approach (future enhancement)
func (ee *ExecutionEngine) executeHybrid(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// For now, default to process execution
	// TODO: Implement smart hybrid execution logic
	return ee.executeProcess(ctx, cmd)
}

// isExternalCommandAvailable checks if an external command exists
func (ee *ExecutionEngine) isExternalCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ExecuteIf executes an if-then-else statement
func (ee *ExecutionEngine) ExecuteIf(ctx context.Context, ifNode *types.IfNode) (*ExecutionResult, error) {
	// Evaluate condition
	conditionResult, err := ee.evaluateCondition(ctx, ifNode.Condition)
	if err != nil {
		return nil, err
	}
	
	// Execute appropriate branch
	if conditionResult {
		return ee.Execute(ctx, ifNode.Then)
	} else if ifNode.Else != nil {
		return ee.Execute(ctx, ifNode.Else)
	}
	
	// No else branch and condition was false
	return &ExecutionResult{
		Success:  true,
		ExitCode: 0,
		Commands: []*CommandResult{},
	}, nil
}

// ExecuteFor executes a for loop
func (ee *ExecutionEngine) ExecuteFor(ctx context.Context, forNode *types.ForNode) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Commands: make([]*CommandResult, 0),
	}
	
	// Iterate over the list
	for _, item := range forNode.List {
		// Set loop variable
		ee.envManager.SetEnv(forNode.Variable, item)
		
		// Execute loop body
		loopResult, err := ee.Execute(ctx, forNode.Body)
		if err != nil {
			return nil, err
		}
		
		result.Commands = append(result.Commands, loopResult.Commands...)
		
		// Check for break/continue (TODO: implement break/continue support)
		if !loopResult.Success {
			result.Success = false
			result.ExitCode = loopResult.ExitCode
			return result, nil
		}
	}
	
	result.Success = true
	result.ExitCode = 0
	return result, nil
}

// ExecuteWhile executes a while loop
func (ee *ExecutionEngine) ExecuteWhile(ctx context.Context, whileNode *types.WhileNode) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Commands: make([]*CommandResult, 0),
	}
	
	maxIterations := 10000 // Safety limit to prevent infinite loops
	iterations := 0
	
	for {
		// Check iteration limit
		if iterations >= maxIterations {
			return nil, fmt.Errorf("while loop exceeded maximum iterations (%d)", maxIterations)
		}
		iterations++
		
		// Evaluate condition
		conditionResult, err := ee.evaluateCondition(ctx, whileNode.Condition)
		if err != nil {
			return nil, err
		}
		
		// Exit loop if condition is false
		if !conditionResult {
			break
		}
		
		// Execute loop body
		loopResult, err := ee.Execute(ctx, whileNode.Body)
		if err != nil {
			return nil, err
		}
		
		result.Commands = append(result.Commands, loopResult.Commands...)
		
		// Check for errors
		if !loopResult.Success {
			result.Success = false
			result.ExitCode = loopResult.ExitCode
			return result, nil
		}
	}
	
	result.Success = true
	result.ExitCode = 0
	return result, nil
}

// evaluateCondition evaluates a condition node and returns true/false
func (ee *ExecutionEngine) evaluateCondition(ctx context.Context, condition types.Node) (bool, error) {
	switch n := condition.(type) {
	case *types.CommandNode:
		// Execute command and check exit code
		cmdResult, err := ee.ExecuteCommand(ctx, n)
		if err != nil {
			return false, err
		}
		return cmdResult.Success && cmdResult.ExitCode == 0, nil
		
	default:
		return false, fmt.Errorf("unsupported condition node type: %T", n)
	}
}

// Helper function to convert error to string
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
