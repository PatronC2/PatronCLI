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

var errUpdateUserUsage = errors.New("update-user usage error")

type updateUserDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

// UpdateUserCommand updates an admin user by username via /api/admin/users/:username.
func UpdateUserCommand(args []string) {
	deps := updateUserDeps{
		GetCreds:    common.GetCreds,
		MakeRequest: common.MakeRequest,
	}
	if err := runUpdateUserCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runUpdateUserCommand(args []string, deps updateUserDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("update-user", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	username := cmd.String("username", "", "Username to update")
	newPassword := cmd.String("new-password", "", "New password")
	newRole := cmd.String("new-role", "", "New role")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errUpdateUserUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set the PATRON_PROFILE environment variable", errUpdateUserUsage)
	}

	trimmedUsername := strings.TrimSpace(*username)
	trimmedRole := strings.TrimSpace(*newRole)
	if trimmedUsername == "" {
		fmt.Fprintln(out, "Missing required field --username")
		cmd.Usage()
		return errUpdateUserUsage
	}
	if *newPassword == "" && trimmedRole == "" {
		fmt.Fprintln(out, "No updates requested. Provide --new-password and/or --new-role")
		cmd.Usage()
		return errUpdateUserUsage
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials. login again or check your profile", selectedProfile)
	}

	reqURL := fmt.Sprintf(
		"https://%s:%s/api/admin/users/%s",
		profile.IP,
		profile.Port,
		url.PathEscape(trimmedUsername),
	)

	requestBody := map[string]interface{}{}
	if *newPassword != "" {
		requestBody["newPassword"] = *newPassword
	}
	if trimmedRole != "" {
		requestBody["newRole"] = trimmedRole
	}

	resp, err := deps.MakeRequest("PUT", reqURL, profile, requestBody)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
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
