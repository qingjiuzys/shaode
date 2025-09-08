package commands

import (
	"gitee.com/com_818cloud/shode/pkg/repl"
	"github.com/spf13/cobra"
)

// NewReplCommand creates the 'repl' command for interactive shell
func NewReplCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repl",
		Short: "Start an interactive shell session",
		Long: `REPL starts an interactive Read-Eval-Print Loop session where you can
execute shell commands in a safe, sandboxed environment with Shode's security features.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create and start the REPL
			shodeRepl := repl.NewREPL()
			shodeRepl.Start()
		},
	}

	return cmd
}
