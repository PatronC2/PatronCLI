package auth

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var errAuthUsage = errors.New("auth usage error")

type authSubcommand struct {
	Execute func(args []string) error
	Help    string
}

// Execute processes the `auth` subcommands.
func Execute(args []string) {
	commands := defaultAuthCommands()
	if err := runAuth(args, commands, os.Stdout); err != nil {
		if errors.Is(err, errAuthUsage) || errors.Is(err, errLoginUsage) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, err)
	}
}

func defaultAuthCommands() map[string]authSubcommand {
	commands := map[string]authSubcommand{}
	commands["configure"] = authSubcommand{
		Execute: func(args []string) error {
			Configure()
			return nil
		},
		Help: "Configure a new profile for authentication.",
	}
	commands["login"] = authSubcommand{
		Execute: func(args []string) error {
			if err := runLoginCommand(args, defaultLoginDeps(), os.Stdin, os.Stdout); err != nil {
				if !errors.Is(err, errLoginUsage) {
					fmt.Fprintln(os.Stdout, err)
				}
				return err
			}
			return nil
		},
		Help: "Login using an existing profile.",
	}
	commands["help"] = authSubcommand{
		Execute: func(args []string) error {
			displayHelp(os.Stdout, commands)
			return nil
		},
		Help: "Show this help menu.",
	}
	return commands
}

func runAuth(args []string, commands map[string]authSubcommand, out io.Writer) error {
	if len(args) < 1 {
		displayHelp(out, commands)
		return errAuthUsage
	}

	commandName := args[0]
	command, exists := commands[commandName]
	if !exists {
		fmt.Fprintf(out, "Unknown auth subcommand: %s\n\n", commandName)
		displayHelp(out, commands)
		return errAuthUsage
	}

	return command.Execute(args[1:])
}

func displayHelp(out io.Writer, commands map[string]authSubcommand) {
	fmt.Fprintln(out, "Usage: patron auth <subcommand> [options]")
	fmt.Fprintln(out, "\nAvailable auth subcommands:")
	for name, cmd := range commands {
		fmt.Fprintf(out, "  %-15s %s\n", name, cmd.Help)
	}
	fmt.Fprintln(out, "\nRun 'patron auth help' for more information about a specific subcommand.")
}
