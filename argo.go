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

func formatCommandValues(values []string) (formatted []string) {

	formatted = make([]string, 0)

	// split a value by `=`
	for _, value := range values {
		if isFlag(value) {
			parts := strings.Split(value, "=")

			for _, part := range parts {
				if strings.Trim(part, " ") != "" {
					formatted = append(formatted, part)
				}
			}
		} else {
			formatted = append(formatted, value)
		}
	}

	return
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

func NewRegistry() Registry {
	return make(Registry)
}

/*
Register will create a new Registry for a base command.
*/
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

/*
So once the user has registered a command we can start
appending them to the registry
*/
func (cr CommandRegistry) AddCommand(cmd *Command) Command {
	cr.Commands[cmd.Name] = cmd
	return *cr.Commands[cmd.Name]
}

func (cmd *Command) AddArg(arg *Arg) (*Arg, bool) {
	// Sanitise inputs
	cleanName := removeWhitespaces(arg.Name)
	cleanDefault := removeWhitespaces(arg.Default)

	variadic := false
	if ok, argName := isVariadicArg(cleanName); ok {
		cleanName = argName
		variadic = true
	}

	if _arg, ok := cmd.Args[cleanName]; ok {
		return _arg, true
	}

	newArg := &Arg{
		Name:       cleanName,
		IsVariadic: variadic,
		Default:    cleanDefault,
	}

	cmd.Args[cleanName] = newArg
	cmd.ArgNames = append(cmd.ArgNames, cleanName)

	return arg, false
}

func (cmd *Command) AddFlag(flg *Flag) (*Flag, bool) {
	// Sanitise inputs
	cleanName := removeWhitespaces(flg.Name)
	cleanShortName := removeWhitespaces(flg.ShortName)
	cleanDefaultValue := removeWhitespaces(flg.Default)

	isInverted := false

	if cleanShortName != "" {
		cleanShortName = cleanShortName[:1]
	}

	if flg.IsBoolean {
		if strings.HasPrefix(flg.Name, "no-") {
			isInverted = true
			cleanName = strings.TrimLeft(flg.Name, "no-")
			cleanDefaultValue = "true"
			cleanShortName = ""
		} else {
			cleanDefaultValue = "false"
		}
	}

	if _flag, ok := cmd.Flags[cleanName]; ok {
		return _flag, true
	}

	flag := &Flag{
		Name:       cleanName,
		ShortName:  cleanShortName,
		IsBoolean:  flg.IsBoolean,
		IsInverted: isInverted,
		Default:    cleanDefaultValue,
	}

	cmd.Flags[cleanName] = flag

	if len(cleanShortName) > 0 {
		cmd.shorthandFlags[cleanShortName] = cleanName
	}

	return flag, false

}

func next(slice []string) (val string, newSlice []string) {
	if len(slice) == 0 {
		val, newSlice = "", make([]string, 0)
		return
	}

	val = slice[0]

	if len(slice) > 1 {
		newSlice = slice[1:]
	} else {
		newSlice = make([]string, 0)
	}

	return
}

func (cmd *CommandRegistry) Parse(args []string) (*Command, error) {
	// Root command should sit at position 0.
	// Maybe there's a better way to do this?
	var commandName string = args[0]

	formattedArguments := formatCommandValues(args)

	// Return ErrorUnsupportedFlag if the flag is unsupported.
	for _, val := range args {
		if isFlag(val) && isUnsupportedFlag(val) {
			return nil, ErrorUnsupportedFlag{val}
		}
	}

	commands := cmd.Commands[commandName]

	for {
		var value string
		value, valuesToParse := next(formattedArguments)

		if len(value) == 0 {
			break
		}

		if isFlag(value) {
			name := strings.TrimLeft(value, "-")
			var flag *Flag

			if isShortFlag(value) {
				if _, ok := commands.shorthandFlags[name]; !ok {
					return nil, ErrorUnknownFlag{value}
				}
			} else {
				if ok, flagName := isInvertedFlag(value); ok {
					if _, ok := commands.Flags[flagName]; !ok {
						return nil, ErrorUnknownFlag{value}
					}

					flag = commands.Flags[flagName]
				} else {
					if _flag, ok := commands.Flags[flagName]; !ok || _flag.IsInverted {
						return nil, ErrorUnknownFlag{value}
					}

					flag = commands.Flags[name]
				}
			}

			if flag.IsBoolean {
				if flag.IsInverted {
					flag.Value = "false"
				} else {
					flag.Value = "true"
				}
			} else {
				if nextValue, nextValuesToProcess := next(valuesToParse); len(nextValue) != 0 && !isFlag(nextValue) {
					flag.Value = nextValue
					valuesToParse = nextValuesToProcess
				}
			}
		} else {
			// process as argument
			for index, argName := range commands.ArgNames {

				// get argument object stored in the `commandConfig`
				arg := commands.Args[argName]

				// assign value if value of the argument is empty
				if len(arg.Value) == 0 {
					arg.Value = value
					break
				}

				// if last argument is a variadic argument, append values
				if (index == len(commands.ArgNames)-1) && arg.IsVariadic {
					arg.Value += fmt.Sprintf(",%s", value)
				}
			}
		}
	}

	return commands, nil
}
