package main

import (
	"fmt"
	"os"
	"patroncli/admin"
	"patroncli/agents"
	"patroncli/auth"
	"patroncli/version"
)

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Printf("patron %s (%s) %s\n", version.Tag, version.Commit, version.Date)
			return
		}
	}

	// existing command map...
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
		"admin": {
			Execute: admin.Execute,
			Help:    "Commands for administrators (e.g. get-log-size, add-user).",
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
	fmt.Println("\nRun 'patron help' to see this message.")
}
