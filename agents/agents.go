package agents

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"patroncli/config"
	"patroncli/types"
)

func Execute(args []string) {
	if len(args) < 1 {
		fmt.Println("agents requires a subcommand like 'list'")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		ListCommand(args[1:])
	default:
		fmt.Printf("unknown agents subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

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

	// Create the HTTP client with a custom Transport to ignore self-signed certificates
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add Authorization header with the profile's token
	req.Header.Set("Authorization", getProfileToken(profile.Name))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch agents, status code: %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Unmarshal the JSON response
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the "agents" field
	agents, ok := response["data"].([]interface{})
	if !ok {
		return fmt.Errorf("response does not contain 'data' field or it is not an array")
	}

	// Apply filtering if specified
	var filteredAgents []interface{}
	for _, agent := range agents {
		agentMap, ok := agent.(map[string]interface{})
		if !ok {
			continue
		}
		if filterKey != "" && filterValue != "" {
			if fmt.Sprintf("%v", agentMap[filterKey]) == filterValue {
				filteredAgents = append(filteredAgents, agent)
			}
		} else {
			filteredAgents = append(filteredAgents, agent)
		}
	}

	// Convert the filtered agents to JSON
	output, err := json.MarshalIndent(filteredAgents, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize agents to JSON: %w", err)
	}

	// Print the JSON output
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
