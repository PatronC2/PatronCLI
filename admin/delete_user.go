package admin

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"patroncli/common"
	"patroncli/types"
	"strings"
)

var errDeleteUserUsage = errors.New("delete-user usage error")

type deleteUserDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

// DeleteUserCommand deletes an admin user by username via /api/admin/users/:username.
func DeleteUserCommand(args []string) {
	deps := deleteUserDeps{
		GetCreds:    common.GetCreds,
		MakeRequest: common.MakeRequest,
	}
	if err := runDeleteUserCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runDeleteUserCommand(args []string, deps deleteUserDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("delete-user", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	username := cmd.String("username", "", "Username to delete")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errDeleteUserUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set the PATRON_PROFILE environment variable", errDeleteUserUsage)
	}

	if strings.TrimSpace(*username) == "" {
		fmt.Fprintln(out, "Missing required field --username")
		cmd.Usage()
		return errDeleteUserUsage
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials. login again or check your profile", selectedProfile)
	}

	reqURL := fmt.Sprintf(
		"https://%s:%s/api/admin/users/%s",
		profile.IP,
		profile.Port,
		url.PathEscape(strings.TrimSpace(*username)),
	)

	resp, err := deps.MakeRequest("DELETE", reqURL, profile, nil)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	var result struct {
		Message string `json:"message"`
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
