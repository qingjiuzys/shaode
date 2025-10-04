# Shode Execution Engine

## Overview

Shode's execution engine provides a complete runtime environment for executing shell scripts with advanced features including pipelines, redirections, control flow, and security sandboxing.

## Features

### 1. Pipeline Support

Execute commands with proper data flow through pipes:

```bash
# Simple pipeline
cat file.txt | grep "pattern" | wc -l

# Complex pipeline with multiple stages
ls -la | awk '{print $9}' | sort | uniq
```

**How it works:**
- Each command's output is fed as input to the next command
- Execution stops if any command fails
- The final command's output is returned

### 2. Input/Output Redirection

Support for all standard redirection operators:

```bash
# Output redirection (overwrite)
echo "Hello World" > output.txt

# Output redirection (append)
echo "More text" >> output.txt

# Input redirection
cat < input.txt

# Error redirection to stdout
command 2>&1

# Redirect both stdout and stderr
command &> all_output.txt
```

### 3. Control Flow

#### If-Then-Else Statements

```bash
if test -f file.txt; then
    echo "File exists"
else
    echo "File not found"
fi
```

#### For Loops

```bash
for item in file1 file2 file3; do
    echo "Processing $item"
    cat "$item"
done
```

#### While Loops

```bash
count=0
while [ $count -lt 5 ]; do
    echo "Count: $count"
    count=$((count + 1))
done
```

**Safety Features:**
- Maximum iteration limit (10,000) to prevent infinite loops
- Context timeout support
- Proper variable scoping

### 4. Variable Assignments

```bash
# Simple assignment
NAME="John"

# Use in commands
echo "Hello, $NAME"

# Environment variable
export PATH="/usr/local/bin:$PATH"
```

### 5. Security Sandbox

All commands are checked against security policies:

**Dangerous Commands Blocked:**
- `rm`, `dd`, `mkfs`, `fdisk` (destructive operations)
- `shutdown`, `reboot`, `halt` (system control)
- `chmod`, `chown`, `passwd` (permission changes)
- `iptables`, `ufw`, `route` (network manipulation)

**Sensitive File Protection:**
- `/etc/passwd`, `/etc/shadow`, `/etc/sudoers`
- `/root/`, `/boot/`, `/dev/`, `/proc/`, `/sys/`

**Pattern Detection:**
- Recursive deletion of root directory
- Password in command line
- Shell injection attempts

## Execution Modes

The engine supports three execution modes:

### 1. Interpreted Mode
For built-in functions and standard library:
```bash
Println "Hello World"
ReadFile "/path/to/file"
WriteFile "/path/to/output" "content"
```

### 2. Process Mode
For external commands:
```bash
ls -la
grep "pattern" file.txt
curl https://example.com
```

### 3. Hybrid Mode
Intelligent switching between interpreted and process modes based on:
- Command availability
- Performance characteristics
- Security requirements

## Using the Execution Engine

### Command Line

#### Run a Script File

```bash
./shode run script.sh
```

Output includes:
- Execution output
- Success/failure status
- Exit code
- Duration
- Number of commands executed

#### Execute Inline Command

```bash
./shode exec "echo Hello World"
```

### Programmatic Usage

```go
import (
    "context"
    "gitee.com/com_818cloud/shode/pkg/engine"
    "gitee.com/com_818cloud/shode/pkg/environment"
    "gitee.com/com_818cloud/shode/pkg/module"
    "gitee.com/com_818cloud/shode/pkg/parser"
    "gitee.com/com_818cloud/shode/pkg/sandbox"
    "gitee.com/com_818cloud/shode/pkg/stdlib"
)

// Parse script
parser := parser.NewSimpleParser()
script, err := parser.ParseFile("script.sh")
if err != nil {
    // Handle error
}

// Create execution engine
envManager := environment.NewEnvironmentManager()
stdLib := stdlib.New()
moduleMgr := module.NewModuleManager()
security := sandbox.NewSecurityChecker()

engine := engine.NewExecutionEngine(envManager, stdLib, moduleMgr, security)

// Execute with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := engine.Execute(ctx, script)
if err != nil {
    // Handle error
}

// Check result
if result.Success {
    fmt.Printf("Execution succeeded in %v\n", result.Duration)
    fmt.Printf("Output: %s\n", result.Output)
} else {
    fmt.Printf("Execution failed with code %d\n", result.ExitCode)
}
```

## Standard Library Functions

Built-in functions that replace common shell commands:

### File System
- `ReadFile(filename)` - Read file contents
- `WriteFile(filename, content)` - Write to file
- `ListFiles(dir)` - List directory contents
- `FileExists(filename)` - Check file existence

### String Operations
- `Contains(haystack, needle)` - String search
- `Replace(str, old, new)` - String replacement
- `ToUpper(str)` - Convert to uppercase
- `ToLower(str)` - Convert to lowercase
- `Trim(str)` - Trim whitespace

### Environment
- `GetEnv(key)` - Get environment variable
- `SetEnv(key, value)` - Set environment variable
- `WorkingDir()` - Get current directory
- `ChangeDir(path)` - Change directory

### Output
- `Print(text)` - Print without newline
- `Println(text)` - Print with newline
- `Error(text)` - Print to stderr
- `Errorln(text)` - Print to stderr with newline

## Performance Features

### Command Caching
- Automatic caching of command results
- Configurable cache size (default: 1000 entries)
- TTL-based expiration
- Cache invalidation support

### Process Pool
- Reusable process pool for repeated commands
- Configurable pool size (default: 10 processes)
- Idle timeout cleanup
- Automatic eviction of old processes

## Best Practices

1. **Use Standard Library** when possible for better performance
2. **Set timeouts** for long-running scripts
3. **Handle errors** properly with if-then-else
4. **Use pipelines** efficiently (avoid unnecessary steps)
5. **Test security** requirements before deployment

## Error Handling

All errors are captured and reported:

```go
type ExecutionResult struct {
    Success  bool
    ExitCode int
    Output   string
    Error    string
    Duration time.Duration
    Commands []*CommandResult
}
```

Individual command results include:
- Command AST node
- Success status
- Exit code
- Output and error messages
- Execution duration
- Execution mode used

## Future Enhancements

- Background job support (`&`)
- Command substitution (`$(...)`)
- Process substitution (`<(...)`)
- Array and associative array support
- Function definitions and calls
- Signal handling
- Debugger integration
