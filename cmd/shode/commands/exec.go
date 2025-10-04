package commands

import (
	"context"
	"fmt"
	"time"

	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"github.com/spf13/cobra"
)

// NewExecCommand creates the 'exec' command for executing inline commands
func NewExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [command]",
		Short: "Execute an inline shell command",
		Long: `Execute runs a single shell command with Shode's security features.
The command will be parsed, analyzed for security risks, and executed in a sandboxed environment.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := args[0]
			if len(args) > 1 {
				// Join all arguments to form the complete command
				for _, arg := range args[1:] {
					command += " " + arg
				}
			}

			fmt.Printf("Executing command: %s\n", command)
			
			// Parse the command
			parser := parser.NewSimpleParser()
			script, err := parser.ParseString(command)
			if err != nil {
				return fmt.Errorf("failed to parse command: %v", err)
			}
			
			if len(script.Nodes) == 0 {
				return fmt.Errorf("no command to execute")
			}
			
			// Initialize execution engine components
			envManager := environment.NewEnvironmentManager()
			stdLib := stdlib.New()
			moduleMgr := module.NewModuleManager()
			security := sandbox.NewSecurityChecker()
			
			// Create execution engine
			executionEngine := engine.NewExecutionEngine(envManager, stdLib, moduleMgr, security)
			
			// Execute the command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()
			
			result, err := executionEngine.Execute(ctx, script)
			if err != nil {
				return fmt.Errorf("execution error: %v", err)
			}
			
			// Display output directly
			if result.Output != "" {
				fmt.Print(result.Output)
			}
			
			// Display errors to stderr
			if result.Error != "" {
				fmt.Fprint(cmd.ErrOrStderr(), result.Error)
			}
			
			// Return error if command failed
			if !result.Success {
				return fmt.Errorf("command execution failed with exit code %d", result.ExitCode)
			}
			
			return nil
		},
	}

	return cmd
}
