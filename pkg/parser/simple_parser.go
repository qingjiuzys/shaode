package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/types"
)

// SimpleParser provides basic shell command parsing without external dependencies
type SimpleParser struct{}

// NewSimpleParser creates a new simple parser
func NewSimpleParser() *SimpleParser {
	return &SimpleParser{}
}

// ParseString parses shell commands from a string
func (p *SimpleParser) ParseString(source string) (*types.ScriptNode, error) {
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
	}

	lines := strings.Split(source, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Simple command parsing
		cmd := p.parseCommand(line, lineNum+1)
		if cmd != nil {
			script.Nodes = append(script.Nodes, cmd)
		}
	}

	return script, nil
}

// ParseFile parses shell commands from a file
func (p *SimpleParser) ParseFile(filename string) (*types.ScriptNode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Simple command parsing
		cmd := p.parseCommand(line, lineNum)
		if cmd != nil {
			script.Nodes = append(script.Nodes, cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return script, nil
}

// parseCommand parses a single line into a command node
func (p *SimpleParser) parseCommand(line string, lineNum int) *types.CommandNode {
	// Simple tokenization - split by spaces, handle quotes
	tokens := p.tokenize(line)
	if len(tokens) == 0 {
		return nil
	}

	cmd := &types.CommandNode{
		Pos: types.Position{
			Line:   lineNum,
			Column: 1,
			Offset: 0,
		},
		Name: tokens[0],
		Args: tokens[1:],
	}

	return cmd
}

// tokenize splits a command line into tokens, handling quotes
func (p *SimpleParser) tokenize(line string) []string {
	var tokens []string
	var currentToken strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(line); i++ {
		char := line[i]

		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				currentToken.WriteByte(char)
			}

		case char == ' ' && !inQuotes:
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}

		default:
			currentToken.WriteByte(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

// DebugPrint prints debug information about parsing
func (p *SimpleParser) DebugPrint(source string) {
	fmt.Println("Simple parser debug output:")
	fmt.Println("Input:", source)

	script, err := p.ParseString(source)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Parsed %d commands:\n", len(script.Nodes))
	for i, node := range script.Nodes {
		if cmd, ok := node.(*types.CommandNode); ok {
			fmt.Printf("  %d: %s %v (line %d)\n", i+1, cmd.Name, cmd.Args, cmd.Pos.Line)
		}
	}
}
