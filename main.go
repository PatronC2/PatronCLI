package main

import (
	"fmt"
	"os"
	"patroncli/agents"
	"patroncli/auth"
)

func main() {
	var commands map[string]struct {
		Execute func(args []string)
		Help    string
	}

	commands = map[string]struct {
		Execute func(args []string)
		Help    string
	}{
		"auth": {
			Execute: auth.Execute,
			Help:    "Commands for authentication (e.g., configure, login).",
		},
		"agents": {
			Execute: agents.Execute,
			Help:    "Commands for managing agents (e.g., list).",
		},
		"help": {
			Execute: func(args []string) { showHelp(commands) },
			Help:    "Show this help menu.",
		},
	}

	if len(os.Args) < 2 {
		showHelp(commands)
		os.Exit(1)
	}

	commandName := os.Args[1]
	command, exists := commands[commandName]
	if !exists {
		fmt.Printf("Unknown command: %s\n\n", commandName)
		showHelp(commands)
		os.Exit(1)
	}

	command.Execute(os.Args[2:])
}

func showHelp(commands map[string]struct {
	Execute func(args []string)
	Help    string
}) {
	fmt.Println("Usage: patron <command> [options]")
	fmt.Println("\nAvailable commands:")
	for name, cmd := range commands {
		fmt.Printf("  %-10s %s\n", name, cmd.Help)
	}
	fmt.Println("\nRun 'patroncli help' to see this message.")
}
