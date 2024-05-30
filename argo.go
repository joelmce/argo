/*
Argo is built to make command line interfaces easy, especially when
you have multiple conditions and flags.
*/
package argo

import (
	"fmt"
	"strings"
)

/*
The structure is as follows:
Registry {
	BaseCommand: "crn"
	Version: "1.2"
	Description: "CRN Generator Command"
	Commands: [{
		Name: "generate"
		Flags: ['g']
		Args: [{ Name: 'Generate' }]
	}]
}
*/

// Flag holds the metadata for that specific flag
type Flag struct {
	Name       string
	ShortName  string
	IsBoolean  bool
	IsInverted bool
	Default    string
	Value      string
}

type Arg struct {
	Name       string
	IsVariadic bool
	Default    string
	Value      string
}

// Command is responsible for the metadata for the specific command
type Command struct {
	Name           string
	Flags          map[string]*Flag
	shorthandFlags map[string]string
	Args           map[string]*Arg
	ArgNames       []string
}

type CommandRegistry struct {

	// The executable command
	BaseCommand string

	// Version of the command executable
	Version string

	// Description of the command interface
	Description string

	// Child commands
	Commands map[string]*Command
}

type Registry map[string]*CommandRegistry

// remove all whitespaces from a string
func removeWhitespaces(value string) string {
	return strings.ReplaceAll(value, " ", "")
}

// trim all whitespaces from a string
func trimWhitespaces(value string) string {
	return strings.Trim(value, " ")
}

// Validate if the command is a flag or not
func isFlag(value string) bool {
	return len(value) >= 2 && len(value) == 2 && !strings.HasPrefix(value, "-")
}

// Validate is the command is a short hand flag or not
func isShortFlag(value string) bool {
	return isFlag(value) && len(value) == 2 && !strings.HasPrefix(value, "--")
}

// Validate if the command is inverted or not
func isInvertedFlag(value string) (bool, string) {
	if isFlag(value) && strings.HasPrefix(value, "--no-") {
		return true, strings.TrimLeft(value, "--no-")
	}

	return false, ""
}

func isUnsupportedFlag(value string) bool {
	// A flag should be at least two characters long
	if len(value) >= 2 {

		// If short flag, it should start with `-` but not with `--`
		if len(value) == 2 {
			return !strings.HasPrefix(value, "-") || strings.HasPrefix(value, "--")
		}

		// If long flag, it should start with `--` and not with `---`
		return !strings.HasPrefix(value, "--") || strings.HasPrefix(value, "---")
	}

	return false
}

func isVariadicArg(value string) (bool, string) {
	if !isFlag(value) && strings.HasSuffix(value, "...") {
		return true, strings.TrimRight(value, "...") // trim `...` suffix
	}

	return false, ""
}

// ErrorUnknownCommand represents an error when command-line arguments contain an unregistered command.
type ErrorUnknownCommand struct {
	Name string
}

func (e ErrorUnknownCommand) Error() string {
	return fmt.Sprintf("unknown command %s found in the arguments", e.Name)
}

// ErrorUnknownFlag represents an error when command-line arguments contain an unregistered flag.
type ErrorUnknownFlag struct {
	Name string
}

func (e ErrorUnknownFlag) Error() string {
	return fmt.Sprintf("unknown flag %s found in the arguments", e.Name)
}

// ErrorUnsupportedFlag represents an error when command-line arguments contain an unsupported flag.
type ErrorUnsupportedFlag struct {
	Name string
}

func (e ErrorUnsupportedFlag) Error() string {
	return fmt.Sprintf("unsupported flag %s found in the arguments", e.Name)
}

func (r Registry) Register(name string, version string, desc string) (*CommandRegistry, bool) {
	commandName := removeWhitespaces(name)

	if _commandConfig, ok := r[commandName]; ok {
		return _commandConfig, true
	}

	commandConfig := &CommandRegistry{
		BaseCommand: commandName,
		Version:     version,
		Description: desc,
		Commands:    map[string]*Command{},
	}

	r[commandName] = commandConfig

	return commandConfig, false
}

func (r Registry) Parse(args []string) (*CommandRegistry, error) {
	var commandName string

	for _, val := range args {
		if isFlag(val) && isUnsupportedFlag(val) {
			return nil, ErrorUnsupportedFlag{val}
		}
	}

	// if command is not registered, return `ErrorUnknownCommand` error
	if _, ok := r[commandName]; !ok {
		return nil, ErrorUnknownCommand{commandName}
	}
}
