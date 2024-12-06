package agents

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"patroncli/common"
)

func ExecCommand(args []string) {
	execCmd := flag.NewFlagSet("exec", flag.ExitOnError)
	profileName := execCmd.String("profile", "", "The profile name to use")
	agentID := execCmd.String("agent-id", "", "The agent ID to execute the command on")
	command := execCmd.String("command", "", "The command to execute")

	execCmd.Parse(args)

	if *profileName == "" {
		fmt.Println("No profile specified. Use --profile flag.")
		execCmd.Usage()
		os.Exit(1)
	}

	if *agentID == "" {
		fmt.Println("No agent specified. Use --agent-id flag.")
		execCmd.Usage()
		os.Exit(1)
	}

	if *command == "" {
		fmt.Println("No command specified. Use --command flag.")
		execCmd.Usage()
		os.Exit(1)
	}

	profile := common.GetCreds(*profileName)

	url := fmt.Sprintf("https://%s:%s/api/command/%s", profile.IP, profile.Port, *agentID)
	requestBody := map[string]string{
		"command": *command,
	}

	responseBody, err := common.MakeRequest("POST", url, profile, requestBody)
	if err != nil {
		fmt.Printf("Error sending command: %v\n", err)
		os.Exit(1)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %v\n", response)
}

func GetCommands(args []string) {
	getCommandsCmd := flag.NewFlagSet("get-responses", flag.ExitOnError)
	profileName := getCommandsCmd.String("profile", "", "The profile name to use")
	agentID := getCommandsCmd.String("agent-id", "", "The agent ID to get responses for")

	getCommandsCmd.Parse(args)

	if *profileName == "" {
		fmt.Println("No profile specified. Use --profile flag.")
		getCommandsCmd.Usage()
		os.Exit(1)
	}

	if *agentID == "" {
		fmt.Println("No agent specified. Use --agent-id flag.")
		getCommandsCmd.Usage()
		os.Exit(1)
	}

	profile := common.GetCreds(*profileName)

	url := fmt.Sprintf("https://%s:%s/api/commands/%s", profile.IP, profile.Port, *agentID)
	responseBody, err := common.MakeRequest("GET", url, profile, nil)
	if err != nil {
		fmt.Printf("Error fetching responses: %v\n", err)
		os.Exit(1)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if len(response.Data) == 0 {
		fmt.Println("No command responses found.")
		return
	}

	fmt.Println("Command Responses:")
	for _, command := range response.Data {
		responseJSON, _ := json.MarshalIndent(command, "", "  ")
		fmt.Println(string(responseJSON))
	}
}
