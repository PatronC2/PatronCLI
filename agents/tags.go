package agents

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"patroncli/common"
	"patroncli/types"
	"strconv"
	"strings"
)

func GetTagsCommand(args []string) {
	cmd := flag.NewFlagSet("get-tags", flag.ExitOnError)
	profileName := cmd.String("profile", "", "The profile name to use")
	agentID := cmd.String("agent-id", "", "The agent UUID to fetch tags for")
	cmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}
	if strings.TrimSpace(*agentID) == "" {
		fmt.Println("No agent specified. Use --agent-id flag.")
		cmd.Usage()
		os.Exit(1)
	}

	profile := common.GetCreds(selectedProfile)
	if err := getTags(profile, strings.TrimSpace(*agentID)); err != nil {
		fmt.Println("Error fetching tags:", err)
		os.Exit(1)
	}
}

func getTags(profile types.Credential, agentID string) error {
	url := fmt.Sprintf("https://%s:%s/api/tags/%s", profile.IP, profile.Port, agentID)
	body, err := common.MakeRequest("GET", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching tags: %w", err)
	}

	var response struct {
		Tags []map[string]interface{} `json:"tags"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize tags to JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func PutTagCommand(args []string) {
	cmd := flag.NewFlagSet("put-tag", flag.ExitOnError)
	profileName := cmd.String("profile", "", "The profile name to use")
	agents := cmd.String("agents", "", "Comma-separated list of agent UUIDs")
	key := cmd.String("key", "", "Tag key")
	value := cmd.String("value", "", "Tag value")
	cmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}
	if strings.TrimSpace(*agents) == "" {
		fmt.Println("No agents specified. Use --agents flag.")
		cmd.Usage()
		os.Exit(1)
	}
	if strings.TrimSpace(*key) == "" {
		fmt.Println("No key specified. Use --key flag.")
		cmd.Usage()
		os.Exit(1)
	}

	agentList := splitCSV(*agents)
	if len(agentList) == 0 {
		fmt.Println("No valid agent IDs found in --agents.")
		os.Exit(1)
	}

	profile := common.GetCreds(selectedProfile)
	if err := putTags(profile, agentList, strings.TrimSpace(*key), *value); err != nil {
		fmt.Println("Error updating tags:", err)
		os.Exit(1)
	}
}

func putTags(profile types.Credential, agents []string, key, value string) error {
	url := fmt.Sprintf("https://%s:%s/api/tag", profile.IP, profile.Port)
	requestBody := map[string]interface{}{
		"agents": agents,
		"key":    key,
		"value":  value,
	}

	body, err := common.MakeRequest("PUT", url, profile, requestBody)
	if err != nil {
		return fmt.Errorf("error updating tags: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize response to JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func DeleteTagCommand(args []string) {
	cmd := flag.NewFlagSet("delete-tag", flag.ExitOnError)
	profileName := cmd.String("profile", "", "The profile name to use")
	tagID := cmd.String("tag-id", "", "Tag ID to delete")
	cmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}
	if strings.TrimSpace(*tagID) == "" {
		fmt.Println("No tag ID specified. Use --tag-id flag.")
		cmd.Usage()
		os.Exit(1)
	}

	id, err := strconv.ParseInt(strings.TrimSpace(*tagID), 10, 64)
	if err != nil || id <= 0 {
		fmt.Println("Invalid tag-id. It must be a positive integer.")
		os.Exit(1)
	}

	profile := common.GetCreds(selectedProfile)
	if err := deleteTag(profile, id); err != nil {
		fmt.Println("Error deleting tag:", err)
		os.Exit(1)
	}
}

func deleteTag(profile types.Credential, tagID int64) error {
	url := fmt.Sprintf("https://%s:%s/api/tag/%d", profile.IP, profile.Port, tagID)
	body, err := common.MakeRequest("DELETE", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error deleting tag: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize response to JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func splitCSV(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}
