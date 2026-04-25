package admin

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"patroncli/common"
	"patroncli/types"
)

var errGetUsersUsage = errors.New("get-users usage error")

type getUsersDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

// GetUsersCommand retrieves users from /api/admin/users.
func GetUsersCommand(args []string) {
	deps := getUsersDeps{
		GetCreds:    common.GetCreds,
		MakeRequest: common.MakeRequest,
	}
	if err := runGetUsersCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runGetUsersCommand(args []string, deps getUsersDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("get-users", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errGetUsersUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set the PATRON_PROFILE environment variable", errGetUsersUsage)
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials. login again or check your profile", selectedProfile)
	}

	reqURL := fmt.Sprintf("https://%s:%s/api/admin/users", profile.IP, profile.Port)
	resp, err := deps.MakeRequest("GET", reqURL, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching users: %w", err)
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	fmt.Fprintln(out, string(output))
	return nil
}
