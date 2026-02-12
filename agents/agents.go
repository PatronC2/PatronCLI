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

	query := searchCmd.String("query", "", "Comma-separated list of fields to include in the output")

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

	var tagList []string
	if strings.TrimSpace(*tags) != "" {
		tagList = strings.Split(*tags, ",")
	}

	err := searchAgents(profile, *hostname, *ip, *status, *logic, tagList, *sort, *limit, *offset, *query)
	if err != nil {
		fmt.Println("Error fetching agents:", err)
		os.Exit(1)
	}
}

func searchAgents(
	profile types.Credential,
	hostname, ip, status, logic string,
	tags []string,
	sort string,
	limit, offset int,
	query string,
) error {
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
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
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

	items := response.Data
	if query != "" {
		items = common.QueryFields(items, query)
	}

	output, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize agents to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
