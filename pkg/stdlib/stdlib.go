package stdlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// StdLib provides built-in functions to replace external commands
type StdLib struct{}

// New creates a new standard library instance
func New() *StdLib {
	return &StdLib{}
}

// FileSystem functions

// ReadFile reads the contents of a file (replaces 'cat')
func (sl *StdLib) ReadFile(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return string(content), nil
}

// WriteFile writes content to a file (replaces echo > file)
func (sl *StdLib) WriteFile(filename, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0644)
}

// ListFiles lists files in a directory (replaces 'ls')
func (sl *StdLib) ListFiles(dirpath string) ([]string, error) {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %v", dirpath, err)
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
}

// FileExists checks if a file exists (replaces test -f)
func (sl *StdLib) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// String functions

// Contains checks if a string contains another string (replaces grep)
func (sl *StdLib) Contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// Replace replaces all occurrences of old with new in a string (replaces sed)
func (sl *StdLib) Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// ToUpper converts string to uppercase (replaces tr '[:lower:]' '[:upper:]')
func (sl *StdLib) ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower converts string to lowercase (replaces tr '[:upper:]' '[:lower:]')
func (sl *StdLib) ToLower(s string) string {
	return strings.ToLower(s)
}

// Trim removes leading and trailing whitespace (replaces sed trimming)
func (sl *StdLib) Trim(s string) string {
	return strings.TrimSpace(s)
}

// Environment functions

// GetEnv gets an environment variable (replaces $VAR)
func (sl *StdLib) GetEnv(key string) string {
	return os.Getenv(key)
}

// SetEnv sets an environment variable (replaces export)
func (sl *StdLib) SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// WorkingDir gets the current working directory (replaces pwd)
func (sl *StdLib) WorkingDir() (string, error) {
	return os.Getwd()
}

// ChangeDir changes the current directory (replaces cd)
func (sl *StdLib) ChangeDir(dirpath string) error {
	return os.Chdir(dirpath)
}

// Utility functions

// Print outputs text to stdout (replaces echo)
func (sl *StdLib) Print(text string) {
	fmt.Print(text)
}

// Println outputs text with newline to stdout (replaces echo)
func (sl *StdLib) Println(text string) {
	fmt.Println(text)
}

// Error outputs text to stderr (replaces echo >&2)
func (sl *StdLib) Error(text string) {
	fmt.Fprint(os.Stderr, text)
}

// Errorln outputs text with newline to stderr (replaces echo >&2)
func (sl *StdLib) Errorln(text string) {
	fmt.Fprintln(os.Stderr, text)
}
