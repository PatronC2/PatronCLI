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

var errSetLogSizeUsage = errors.New("set-log-size usage error")

type setLogSizeDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

// SetLogSizeCommand updates max log file size for an app via /api/admin/log-size.
func SetLogSizeCommand(args []string) {
	deps := setLogSizeDeps{
		GetCreds:    common.GetCreds,
		MakeRequest: common.MakeRequest,
	}
	if err := runSetLogSizeCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runSetLogSizeCommand(args []string, deps setLogSizeDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("set-log-size", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	appName := cmd.String("app", "", "App name (e.g. api, server)")
	size := cmd.Int64("size", 0, "Max log file size value (positive integer)")
	unit := cmd.String("unit", "", "Size unit (MB or GB)")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errSetLogSizeUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}

	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set the PATRON_PROFILE environment variable", errSetLogSizeUsage)
	}

	normalizedUnit := strings.ToUpper(strings.TrimSpace(*unit))
	if *appName == "" || *size <= 0 || normalizedUnit == "" {
		fmt.Fprintln(out, "Missing or invalid fields. Required: --app, --size (>0), --unit (MB|GB)")
		cmd.Usage()
		return errSetLogSizeUsage
	}
	if normalizedUnit != "MB" && normalizedUnit != "GB" {
		return fmt.Errorf("unit must be MB or GB")
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials. login again or check your profile", selectedProfile)
	}

	reqURL := fmt.Sprintf("https://%s:%s/api/admin/log-size", profile.IP, profile.Port)
	requestBody := map[string]interface{}{
		"app":  *appName,
		"size": *size,
		"unit": normalizedUnit,
	}

	resp, err := deps.MakeRequest("PUT", reqURL, profile, requestBody)
	if err != nil {
		return fmt.Errorf("error setting log size: %w", err)
	}

	var result struct {
		Message   string `json:"message"`
		App       string `json:"app"`
		SizeBytes int64  `json:"size_bytes"`
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
