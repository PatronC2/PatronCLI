package redirectors

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

var errGetRedirectorsUsage = errors.New("get-redirectors usage error")
var errCreateRedirectorUsage = errors.New("create-redirector usage error")

type redirectorDeps struct {
	GetCreds    func(profile string) types.Credential
	MakeRequest func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
}

// GetRedirectorsCommand fetches redirectors from /api/redirectors.
func GetRedirectorsCommand(args []string) {
	deps := redirectorDeps{GetCreds: common.GetCreds, MakeRequest: common.MakeRequest}
	if err := runGetRedirectorsCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runGetRedirectorsCommand(args []string, deps redirectorDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("get-redirectors", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	filter := cmd.String("filter", "", "Filter expression (e.g. name=edge-a)")
	query := cmd.String("query", "", "Comma-separated fields to return")
	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errGetRedirectorsUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set PATRON_PROFILE", errGetRedirectorsUsage)
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials", selectedProfile)
	}

	url := fmt.Sprintf("https://%s:%s/api/redirectors", profile.IP, profile.Port)
	resp, err := deps.MakeRequest("GET", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error fetching redirectors: %w", err)
	}

	var parsed struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp, &parsed); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	items := parsed.Data
	if strings.TrimSpace(*filter) != "" {
		items = common.FilterItemsWithTags(items, *filter)
	}
	if strings.TrimSpace(*query) != "" {
		items = common.QueryFields(items, *query)
	}
	parsed.Data = items

	output, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}
	fmt.Fprintln(out, string(output))
	return nil
}

// CreateRedirectorCommand creates a redirector via /api/redirector.
func CreateRedirectorCommand(args []string) {
	deps := redirectorDeps{GetCreds: common.GetCreds, MakeRequest: common.MakeRequest}
	if err := runCreateRedirectorCommand(args, deps, os.Stdout); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func runCreateRedirectorCommand(args []string, deps redirectorDeps, out io.Writer) error {
	cmd := flag.NewFlagSet("create-redirector", flag.ContinueOnError)
	cmd.SetOutput(out)
	profileName := cmd.String("profile", "", "The profile name to use")
	name := cmd.String("name", "", "Redirector name")
	description := cmd.String("description", "", "Redirector description")
	forwardIP := cmd.String("forward-ip", "", "Forward target IPv4")
	forwardPort := cmd.String("forward-port", "", "Forward target port")
	listenIPv4 := cmd.String("listen-ipv4", "", "Redirector listen IPv4")
	listenIPv6 := cmd.String("listen-ipv6", "", "Redirector listen IPv6 (optional)")
	listenPort := cmd.String("listen-port", "", "External listen port")

	if err := cmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errCreateRedirectorUsage, err)
	}

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		return fmt.Errorf("%w: no profile specified. use --profile flag or set PATRON_PROFILE", errCreateRedirectorUsage)
	}

	if strings.TrimSpace(*name) == "" || strings.TrimSpace(*forwardIP) == "" || strings.TrimSpace(*forwardPort) == "" || strings.TrimSpace(*listenIPv4) == "" || strings.TrimSpace(*listenPort) == "" {
		fmt.Fprintln(out, "Missing required fields. Required: --name, --forward-ip, --forward-port, --listen-ipv4, --listen-port")
		cmd.Usage()
		return errCreateRedirectorUsage
	}

	profile := deps.GetCreds(selectedProfile)
	if profile.IP == "" || profile.Port == "" {
		return fmt.Errorf("profile %q is missing IP/Port credentials", selectedProfile)
	}

	url := fmt.Sprintf("https://%s:%s/api/redirector", profile.IP, profile.Port)
	body := map[string]string{
		"Name":        strings.TrimSpace(*name),
		"Description": strings.TrimSpace(*description),
		"ForwardIP":   strings.TrimSpace(*forwardIP),
		"ForwardPort": strings.TrimSpace(*forwardPort),
		"ListenIPv4":  strings.TrimSpace(*listenIPv4),
		"ListenIPv6":  strings.TrimSpace(*listenIPv6),
		"ListenPort":  strings.TrimSpace(*listenPort),
	}

	resp, err := deps.MakeRequest("POST", url, profile, body)
	if err != nil {
		return fmt.Errorf("error creating redirector: %w", err)
	}

	// API returns shell script bytes on success. Print raw output.
	fmt.Fprintln(out, string(resp))
	return nil
}
