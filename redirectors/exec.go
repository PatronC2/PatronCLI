package redirectors

import (
	"fmt"
	"os"
)

func Execute(args []string) {
	var commands map[string]struct {
		Execute func(args []string)
		Help    string
	}

	commands = map[string]struct {
		Execute func(args []string)
		Help    string
	}{
		"list": {
			Execute: GetRedirectorsCommand,
			Help:    "List redirectors.",
		},
		"create": {
			Execute: CreateRedirectorCommand,
			Help:    "Create a redirector.",
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
		fmt.Printf("Unknown redirectors subcommand: %s\n\n", commandName)
		displayHelp(commands)
		os.Exit(1)
	}

	command.Execute(args[1:])
}

func displayHelp(commands map[string]struct {
	Execute func(args []string)
	Help    string
}) {
	fmt.Println("Usage: patron redirectors <subcommand> [options]")
	fmt.Println("\nAvailable redirectors subcommands:")
	for name, cmd := range commands {
		fmt.Printf("  %-15s %s\n", name, cmd.Help)
	}
	fmt.Println("\nRun 'patron redirectors help' for more information about a specific subcommand.")
}
