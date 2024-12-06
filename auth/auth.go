package auth

import (
	"fmt"
	"os"
)

// Execute processes the `auth` subcommands
func Execute(args []string) {
	var commands map[string]struct {
		Execute func(args []string)
		Help    string
	}

	commands = map[string]struct {
		Execute func(args []string)
		Help    string
	}{
		"configure": {
			Execute: func(args []string) { Configure() },
			Help:    "Configure a new profile for authentication.",
		},
		"login": {
			Execute: LoginCommand,
			Help:    "Login using an existing profile.",
		},
		"help": {
			Execute: func(args []string) { displayHelp(commands) },
			Help:    "Show this help menu.",
		},
	}

	if len(args) < 1 {
		displayHelp(commands)
		os.Exit(1)
	}

	commandName := args[0]
	command, exists := commands[commandName]
	if !exists {
		fmt.Printf("Unknown auth subcommand: %s\n\n", commandName)
		displayHelp(commands)
		os.Exit(1)
	}

	command.Execute(args[1:])
}

func displayHelp(commands map[string]struct {
	Execute func(args []string)
	Help    string
}) {
	fmt.Println("Usage: patron auth <subcommand> [options]")
	fmt.Println("\nAvailable auth subcommands:")
	for name, cmd := range commands {
		fmt.Printf("  %-15s %s\n", name, cmd.Help)
	}
	fmt.Println("\nRun 'patron auth help' for more information about a specific subcommand.")
}
