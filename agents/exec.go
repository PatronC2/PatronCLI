package agents

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
			Help:    "Search agents, optionally filtering by criteria.",
		},
		"describe": {
			Execute: DescribeCommand,
			Help:    "Describe a specific agent using --agent-id.",
		},
		"send-command": {
			Execute: ExecCommand,
			Help:    "Send a command to an agent using --agent-id and --command.",
		},
		"get-commands": {
			Execute: GetCommands,
			Help:    "Retrieve command responses from an agent using --agent-id.",
		},
		"get-tags": {
			Execute: GetTagsCommand,
			Help:    "Get tags for an agent using --agent-id.",
		},
		"put-tag": {
			Execute: PutTagCommand,
			Help:    "Set/update a tag for one or more agents using --agents, --key, --value.",
		},
		"delete-tag": {
			Execute: DeleteTagCommand,
			Help:    "Delete a tag by ID using --tag-id.",
		},
		"list-files": {
			Execute: ListFilesCommand,
			Help:    "List files for an agent using --agent-id.",
		},
		"download-file": {
			Execute: DownloadFileCommand,
			Help:    "Download a file by ID to local disk using --file-id and --output.",
		},
		"upload-file": {
			Execute: UploadFileCommand,
			Help:    "Upload/queue a file operation using --agent-id, --path, --transfer-type, [--source].",
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
