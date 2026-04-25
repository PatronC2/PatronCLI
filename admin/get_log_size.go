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

var errGetLogSizeUsage = errors.New("get-log-size usage error")

type getLogSizeDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

// GetLogSizeCommand retrieves max log size for a specific app from /api/admin/log-size.
func GetLogSizeCommand(args []string) {
	deps := getLogSizeDeps{
		GetCreds:    common.GetCreds,
		MakeRequest: common.MakeRequest,
	}
	if err := runGetLogSizeCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runGetLogSizeCommand(args []string, deps getLogSizeDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("get-log-size", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	appName := cmd.String("app", "", "App name (e.g. api, server)")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errGetLogSizeUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set the PATRON_PROFILE environment variable", errGetLogSizeUsage)
	}

	if *appName == "" {
		fmt.Fprintln(out, "Missing required --app flag.")
		cmd.Usage()
		return errGetLogSizeUsage
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials. login again or check your profile", selectedProfile)
	}

	reqURL := fmt.Sprintf(
		"https://%s:%s/api/admin/log-size?app=%s",
		profile.IP,
		profile.Port,
		url.QueryEscape(*appName),
	)

	resp, err := deps.MakeRequest("GET", reqURL, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching log size: %w", err)
	}

	var result struct {
		App           string `json:"app"`
		HumanReadable string `json:"human_readable"`
		SizeBytes     int64  `json:"size_bytes"`
		SizeMB        int64  `json:"size_mb"`
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
