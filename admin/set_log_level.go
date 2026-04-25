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
)

var errSetLogLevelUsage = errors.New("set-log-level usage error")

type setLogLevelDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

var validLogLevels = map[string]bool{
	"debug":   true,
	"info":    true,
	"warning": true,
	"error":   true,
}

// SetLogLevelCommand updates log level for a specific app via /api/admin/logging.
func SetLogLevelCommand(args []string) {
	deps := setLogLevelDeps{
		GetCreds:    common.GetCreds,
		MakeRequest: common.MakeRequest,
	}
	if err := runSetLogLevelCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runSetLogLevelCommand(args []string, deps setLogLevelDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("set-log-level", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	appName := cmd.String("app", "", "App name (e.g. api, server)")
	logLevel := cmd.String("log-level", "", "Log level (debug, info, warning, error)")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errSetLogLevelUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set the PATRON_PROFILE environment variable", errSetLogLevelUsage)
	}

	if *appName == "" || *logLevel == "" {
		fmt.Fprintln(out, "Missing required query parameters: app and log_level")
		cmd.Usage()
		return errSetLogLevelUsage
	}

	if !validLogLevels[*logLevel] {
		return fmt.Errorf("invalid log level %q. allowed values: debug, info, warning, error", *logLevel)
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials. login again or check your profile", selectedProfile)
	}

	reqURL := fmt.Sprintf(
		"https://%s:%s/api/admin/logging?app=%s&log_level=%s",
		profile.IP,
		profile.Port,
		url.QueryEscape(*appName),
		url.QueryEscape(*logLevel),
	)

	resp, err := deps.MakeRequest("PUT", reqURL, profile, nil)
	if err != nil {
		return fmt.Errorf("error setting log level: %w", err)
	}

	var result struct {
		Message  string `json:"message"`
		App      string `json:"app"`
		LogLevel string `json:"log_level"`
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
