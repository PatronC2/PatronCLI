package files

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
		"search": {
			Execute: SearchCommand,
			Help:    "Search files, optionally filtering by tags.",
		},
		"download": {
			Execute: DownloadFileCommand,
			Help:    "Download a file by ID to local disk using --file-id and --output.",
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
		fmt.Printf("Unknown files subcommand: %s\n\n", commandName)
		displayHelp(commands)
		os.Exit(1)
	}

	command.Execute(args[1:])
}

func displayHelp(commands map[string]struct {
	Execute func(args []string)
	Help    string
}) {
	fmt.Println("Usage: patron files <subcommand> [options]")
	fmt.Println("\nAvailable files subcommands:")
	for name, cmd := range commands {
		fmt.Printf("  %-15s %s\n", name, cmd.Help)
	}
	fmt.Println("\nRun 'patron files help' for more information about a specific subcommand.")
}
