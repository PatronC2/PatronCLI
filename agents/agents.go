package agents

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"patroncli/common"
	"patroncli/types"
	"strconv"
	"strings"
)

func ListCommand(args []string) {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	profileName := listCmd.String("profile", "", "The profile name to use")
	filter := listCmd.String("filter", "", "Filter agents by criteria (e.g., 'tags.key=value')")
	query := listCmd.String("query", "", "Comma-separated list of fields to include in the output")

	listCmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}

	profile := common.GetCreds(selectedProfile)

	err := fetchAgents(profile, *filter, *query)
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
	profile := common.GetCreds(selectedProfile)
	// Make the GET request to /api/agent with filtering
	err := describeAgent(profile, *agentId)
	if err != nil {
		fmt.Println("Error fetching agent:", err)
		os.Exit(1)
	}
}

func fetchAgents(profile types.Credential, filter, query string) error {
	url := fmt.Sprintf("https://%s:%s/api/agents", profile.IP, profile.Port)

	body, err := common.MakeRequest("GET", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching agents: %w", err)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	filteredAgents := common.FilterItemsWithTags(response.Data, filter)

	if query != "" {
		filteredAgents = common.QueryFields(filteredAgents, query)
	}

	output, err := json.MarshalIndent(filteredAgents, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize agents to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func describeAgent(profile types.Credential, agentId string) error {
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

func SearchCommand(args []string) {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	profileName := searchCmd.String("profile", "", "The profile name to use")
	hostname := searchCmd.String("hostname", "", "Hostname filter")
	ip := searchCmd.String("ip", "", "IP filter")
	status := searchCmd.String("status", "", "Status filter (Online/Offline)")
	logic := searchCmd.String("logic", "or", "Tag logic (or/and)")
	tags := searchCmd.String("tags", "", "Comma-separated tag filters (e.g. key:value,key2:value2)")
	sort := searchCmd.String("sort", "", "Sort field (e.g. hostname:asc)")
	limit := searchCmd.Int("limit", 20, "Number of results")
	offset := searchCmd.Int("offset", 0, "Offset for pagination")

	searchCmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}

	profile := common.GetCreds(selectedProfile)

	tagList := strings.Split(*tags, ",")
	err := searchAgents(profile, *hostname, *ip, *status, *logic, tagList, *sort, *limit, *offset)

	if err != nil {
		fmt.Println("Error fetching agents:", err)
		os.Exit(1)
	}
}

func searchAgents(profile types.Credential, hostname, ip, status, logic string, tags []string, sort string, limit, offset int) error {
	baseUrl := fmt.Sprintf("https://%s:%s/api/agents/search", profile.IP, profile.Port)
	queryParams := url.Values{}

	if hostname != "" {
		queryParams.Add("hostname", hostname)
	}
	if ip != "" {
		queryParams.Add("ip", ip)
	}
	if status != "" {
		queryParams.Add("status", status)
	}
	if logic != "" {
		queryParams.Add("logic", logic)
	}
	if sort != "" {
		queryParams.Add("sort", sort)
	}
	if limit > 0 {
		queryParams.Add("limit", strconv.Itoa(limit))
	}
	if offset >= 0 {
		queryParams.Add("offset", strconv.Itoa(offset))
	}
	for _, tag := range tags {
		queryParams.Add("tag", tag)
	}

	fullUrl := baseUrl + "?" + queryParams.Encode()

	body, err := common.MakeRequest("GET", fullUrl, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching agents: %w", err)
	}

	var response struct {
		Data       []map[string]interface{} `json:"data"`
		TotalCount int                      `json:"totalCount"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	output, err := json.MarshalIndent(response.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize agents to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
