package admin

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunGetLogSizeCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotProfileName string
	var gotMethod string
	var gotURL string

	deps := getLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "172.16.200.185", Port: "8443", Token: "Bearer x"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			return []byte(`{"app":"api","human_readable":"10 MB","size_bytes":10485760,"size_mb":10}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetLogSizeCommand([]string{"--profile", "prod", "--app", "api"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotProfileName != "prod" {
		t.Fatalf("expected profile 'prod', got %q", gotProfileName)
	}
	if gotMethod != "GET" {
		t.Fatalf("expected GET method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/log-size?app=api" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}

	output := out.String()
	if !strings.Contains(output, "\"app\": \"api\"") || !strings.Contains(output, "\"size_mb\": 10") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunGetLogSizeCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfileName string
	deps := getLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"app":"server","human_readable":"10 MB","size_bytes":10485760,"size_mb":10}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetLogSizeCommand([]string{"--app", "server"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfileName != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfileName)
	}
}

func TestRunGetLogSizeCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogSizeDeps{}
	var out bytes.Buffer
	err := runGetLogSizeCommand([]string{"--app", "api"}, deps, &out)
	if !errors.Is(err, errGetLogSizeUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunGetLogSizeCommand_MissingApp(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := getLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runGetLogSizeCommand([]string{"--profile", "p1"}, deps, &out)
	if !errors.Is(err, errGetLogSizeUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
	if !strings.Contains(out.String(), "Missing required --app flag") {
		t.Fatalf("expected app usage message, got %q", out.String())
	}
}

func TestRunGetLogSizeCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{}
		},
	}

	var out bytes.Buffer
	err := runGetLogSizeCommand([]string{"--profile", "p1", "--app", "api"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunGetLogSizeCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runGetLogSizeCommand([]string{"--profile", "p1", "--app", "api"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error fetching log size") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunGetLogSizeCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runGetLogSizeCommand([]string{"--profile", "p1", "--app", "api"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
