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
	"strings"
)

var errCreateUserUsage = errors.New("create-user usage error")

type createUserDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

// CreateUserCommand creates an admin user via /api/admin/users.
func CreateUserCommand(args []string) {
	deps := createUserDeps{
		GetCreds:    common.GetCreds,
		MakeRequest: common.MakeRequest,
	}
	if err := runCreateUserCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runCreateUserCommand(args []string, deps createUserDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("create-user", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	username := cmd.String("username", "", "Username for the new user")
	role := cmd.String("role", "", "Role for the new user")
	password := cmd.String("password", "", "Password for the new user")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errCreateUserUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set the PATRON_PROFILE environment variable", errCreateUserUsage)
	}

	if strings.TrimSpace(*username) == "" || strings.TrimSpace(*role) == "" || *password == "" {
		fmt.Fprintln(out, "Missing required fields. Required: --username, --role, --password")
		cmd.Usage()
		return errCreateUserUsage
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials. login again or check your profile", selectedProfile)
	}

	reqURL := fmt.Sprintf("https://%s:%s/api/admin/users", profile.IP, profile.Port)
	requestBody := map[string]interface{}{
		"username":        strings.TrimSpace(*username),
		"role":            strings.TrimSpace(*role),
		"password":        *password,
		"confirmPassword": *password,
	}

	resp, err := deps.MakeRequest("POST", reqURL, profile, requestBody)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
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
