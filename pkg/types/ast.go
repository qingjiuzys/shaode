package types

// Node represents a generic AST node
type Node interface {
	Position() Position
	String() string
}

// Position represents the source code position of a node
type Position struct {
	Line   int
	Column int
	Offset int
}

// CommandNode represents a shell command
type CommandNode struct {
	Pos      Position
	Name     string
	Args     []string
	Redirect *RedirectNode
}

func (n *CommandNode) Position() Position { return n.Pos }
func (n *CommandNode) String() string     { return n.Name }

// PipeNode represents a pipe between commands
type PipeNode struct {
	Pos  Position
	Left Node
	Right Node
}

func (n *PipeNode) Position() Position { return n.Pos }
func (n *PipeNode) String() string     { return "|" }

// RedirectNode represents input/output redirection
type RedirectNode struct {
	Pos   Position
	Op    string // >, >>, <, etc.
	File  string
	Fd    int // file descriptor (0, 1, 2)
}

func (n *RedirectNode) Position() Position { return n.Pos }
func (n *RedirectNode) String() string     { return n.Op }

// ScriptNode represents a complete shell script
type ScriptNode struct {
	Pos   Position
	Nodes []Node
}

func (n *ScriptNode) Position() Position { return n.Pos }
func (n *ScriptNode) String() string     { return "script" }
