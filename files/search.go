package files

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
	logic := searchCmd.String("logic", "or", "Tag logic (or/and)")
	tags := searchCmd.String("tags", "", "Comma-separated tag filters (e.g. key:value,key2:value2)")
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

	err := searchFiles(profile, *logic, tagList, *limit, *offset, *query)
	if err != nil {
		fmt.Println("Error fetching files:", err)
		os.Exit(1)
	}
}

func searchFiles(profile types.Credential, logic string, tags []string, limit, offset int, query string) error {
	baseURL := fmt.Sprintf("https://%s:%s/api/files/list", profile.IP, profile.Port)
	queryParams := url.Values{}

	if logic != "" {
		queryParams.Add("logic", logic)
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

	fullURL := baseURL
	if encoded := queryParams.Encode(); encoded != "" {
		fullURL += "?" + encoded
	}

	body, err := common.MakeRequest("GET", fullURL, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching files: %w", err)
	}

	var response struct {
		Data       []map[string]interface{} `json:"data"`
		TotalCount int                      `json:"totalCount"`
		NextOffset int                      `json:"nextOffset"`
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
		return fmt.Errorf("failed to serialize files to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
