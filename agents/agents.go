package agents

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"patroncli/common"
	"patroncli/config"
	"patroncli/types"
)

func ListCommand(args []string) {
	// Create a flag set for the `list` subcommand
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	profileName := listCmd.String("profile", "", "The profile name to use")
	filterKey := listCmd.String("filter-key", "", "The key to filter agents by")
	filterValue := listCmd.String("filter-value", "", "The value to filter agents by")

	// Parse the flags
	listCmd.Parse(args)

	// Determine the profile to use
	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}

	// Fetch the profile details
	profile, err := getProfile(selectedProfile)
	if err != nil {
		fmt.Println("Error fetching profile:", err)
		os.Exit(1)
	}

	// Make the GET request to /api/agents with filtering
	err = fetchAgents(profile, *filterKey, *filterValue)
	if err != nil {
		fmt.Println("Error fetching agents:", err)
		os.Exit(1)
	}
}

func DescribeCommand(args []string) {
	// Create a flag set for the `describe` subcommand
	describeCmd := flag.NewFlagSet("describe", flag.ExitOnError)
	profileName := describeCmd.String("profile", "", "The profile name to use")
	agentId := describeCmd.String("agent-id", "", "The agent ID to get")

	// Parse the flags
	describeCmd.Parse(args)

	// Determine the profile to use
	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}

	if *agentId == "" {
		fmt.Println("No agent specified. Use --agent-id")
		os.Exit(1)
	}

	// Fetch the profile details
	profile, err := getProfile(selectedProfile)
	if err != nil {
		fmt.Println("Error fetching profile:", err)
		os.Exit(1)
	}

	// Make the GET request to /api/agent with filtering
	err = describeAgent(profile, *agentId)
	if err != nil {
		fmt.Println("Error fetching agent:", err)
		os.Exit(1)
	}
}

func getProfile(profileName string) (types.Profile, error) {
	profilesPath := config.GetConfigPath()
	data, err := os.ReadFile(profilesPath)
	if err != nil {
		return types.Profile{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var profiles []types.Profile
	err = json.Unmarshal(data, &profiles)
	if err != nil {
		return types.Profile{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	for _, profile := range profiles {
		if profile.Name == profileName {
			return profile, nil
		}
	}

	return types.Profile{}, fmt.Errorf("profile '%s' not found", profileName)
}

func fetchAgents(profile types.Profile, filterKey, filterValue string) error {
	url := fmt.Sprintf("https://%s:%s/api/agents", profile.IP, profile.Port)

	body, err := common.MakeRequest("GET", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching agents: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	agents, ok := response["data"].([]interface{})
	if !ok {
		return fmt.Errorf("response does not contain 'data' field or it is not an array")
	}

	filteredAgents := common.FilterItems(agents, filterKey, filterValue)

	output, err := json.MarshalIndent(filteredAgents, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize agents to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func describeAgent(profile types.Profile, agentId string) error {
	url := fmt.Sprintf("https://%s:%s/api/agent/%s", profile.IP, profile.Port, agentId)

	body, err := common.MakeRequest("GET", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching agent: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	agent, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("response does not contain 'data' field or it is not an object")
	}

	output, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize agent to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func getProfileToken(profileName string) string {
	credentialsPath := config.GetCredentialsPath()
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		fmt.Printf("Warning: failed to read credentials file: %v\n", err)
		return ""
	}

	var credentials []types.Credential
	_ = json.Unmarshal(data, &credentials)

	for _, cred := range credentials {
		if cred.Profile == profileName {
			return cred.Token
		}
	}

	fmt.Printf("Warning: token not found for profile '%s'\n", profileName)
	return ""
}
