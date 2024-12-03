package auth

import (
	"fmt"
	"os"
)

// Execute processes the `auth` subcommands
func Execute(args []string) {
	if len(args) < 1 {
		fmt.Println("auth requires a subcommand like 'configure' or 'login'")
		os.Exit(1)
	}

	switch args[0] {
	case "configure":
		// Handle "auth configure"
		Configure()
	case "login":
		// Handle "auth login"
		LoginCommand(args[1:])
	default:
		fmt.Printf("unknown auth subcommand: %s\n", args[0])
		os.Exit(1)
	}
}
