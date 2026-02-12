package admin

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
		"get-log-size": {
			Execute: GetLogSizeCommand,
			Help:    "Get max log file size for an app (use --app).",
		},
		"get-log-level": {
			Execute: GetLogLevelCommand,
			Help:    "Get current log level for an app (use --app).",
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
		fmt.Printf("Unknown agents subcommand: %s\n\n", commandName)
		displayHelp(commands)
		os.Exit(1)
	}

	command.Execute(args[1:])
}

func displayHelp(commands map[string]struct {
	Execute func(args []string)
	Help    string
}) {
	fmt.Println("Usage: patron agents <subcommand> [options]")
	fmt.Println("\nAvailable agents subcommands:")
	for name, cmd := range commands {
		fmt.Printf("  %-15s %s\n", name, cmd.Help)
	}
	fmt.Println("\nRun 'patron agents help' for more information about a specific subcommand.")
}
