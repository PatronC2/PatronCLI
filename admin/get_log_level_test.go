package admin

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunGetLogLevelCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotProfileName string
	var gotMethod string
	var gotURL string

	deps := getLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "172.16.200.185", Port: "8443", Token: "Bearer x"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			return []byte(`{"app":"api","log_level":"warning"}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetLogLevelCommand([]string{"--profile", "prod", "--app", "api"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotProfileName != "prod" {
		t.Fatalf("expected profile 'prod', got %q", gotProfileName)
	}
	if gotMethod != "GET" {
		t.Fatalf("expected GET method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/logging?app=api" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}

	output := out.String()
	if !strings.Contains(output, "\"app\": \"api\"") || !strings.Contains(output, "\"log_level\": \"warning\"") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunGetLogLevelCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfileName string
	deps := getLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"app":"server","log_level":"info"}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetLogLevelCommand([]string{"--app", "server"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfileName != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfileName)
	}
}

func TestRunGetLogLevelCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogLevelDeps{}
	var out bytes.Buffer
	err := runGetLogLevelCommand([]string{"--app", "api"}, deps, &out)
	if !errors.Is(err, errGetLogLevelUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunGetLogLevelCommand_MissingApp(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := getLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runGetLogLevelCommand([]string{"--profile", "p1"}, deps, &out)
	if !errors.Is(err, errGetLogLevelUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
	if !strings.Contains(out.String(), "Missing required --app flag") {
		t.Fatalf("expected app usage message, got %q", out.String())
	}
}

func TestRunGetLogLevelCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{}
		},
	}

	var out bytes.Buffer
	err := runGetLogLevelCommand([]string{"--profile", "p1", "--app", "api"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunGetLogLevelCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runGetLogLevelCommand([]string{"--profile", "p1", "--app", "api"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error fetching log level") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunGetLogLevelCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runGetLogLevelCommand([]string{"--profile", "p1", "--app", "api"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
