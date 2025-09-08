package commands

import (
	"fmt"
	"os"

	"gitee.com/com_818cloud/shode/pkg/parser"
	"github.com/spf13/cobra"
)

// NewRunCommand creates the 'run' command for executing script files
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [script-file]",
		Short: "Run a shell script file",
		Long: `Run executes a shell script file with Shode's security features enabled.
The script will be parsed, analyzed for security risks, and executed in a sandboxed environment.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptFile := args[0]
			
			// Check if file exists
			if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
				return fmt.Errorf("script file not found: %s", scriptFile)
			}

			fmt.Printf("Running script: %s\n", scriptFile)
			
			// Parse the script file
			parser := parser.NewSimpleParser()
			script, err := parser.ParseFile(scriptFile)
			if err != nil {
				return fmt.Errorf("failed to parse script: %v", err)
			}
			
			fmt.Printf("Parsed %d commands successfully\n", len(script.Nodes))
			fmt.Println("(Shode execution engine will execute the commands here)")
			
			// TODO: Implement execution engine and security checks
			
			return nil
		},
	}

	return cmd
}
