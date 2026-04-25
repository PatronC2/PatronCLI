package admin

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunSetLogLevelCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotProfileName string
	var gotMethod string
	var gotURL string

	deps := setLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "172.16.200.185", Port: "8443", Token: "Bearer x"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			return []byte(`{"app":"api","log_level":"info","message":"Log level updated"}`), nil
		},
	}

	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--profile", "prod", "--app", "api", "--log-level", "info"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotProfileName != "prod" {
		t.Fatalf("expected profile 'prod', got %q", gotProfileName)
	}
	if gotMethod != "PUT" {
		t.Fatalf("expected PUT method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/logging?app=api&log_level=info" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}

	output := out.String()
	if !strings.Contains(output, "\"log_level\": \"info\"") || !strings.Contains(output, "\"message\": \"Log level updated\"") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunSetLogLevelCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfileName string
	deps := setLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"app":"server","log_level":"warning","message":"Log level updated"}`), nil
		},
	}

	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--app", "server", "--log-level", "warning"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfileName != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfileName)
	}
}

func TestRunSetLogLevelCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogLevelDeps{}
	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--app", "api", "--log-level", "info"}, deps, &out)
	if !errors.Is(err, errSetLogLevelUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunSetLogLevelCommand_MissingRequiredFlags(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := setLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--profile", "p1", "--app", "api"}, deps, &out)
	if !errors.Is(err, errSetLogLevelUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
	if !strings.Contains(out.String(), "Missing required query parameters") {
		t.Fatalf("expected usage message, got %q", out.String())
	}
}

func TestRunSetLogLevelCommand_InvalidLogLevel(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--profile", "p1", "--app", "api", "--log-level", "trace"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "invalid log level") {
		t.Fatalf("expected invalid log level error, got %v", err)
	}
}

func TestRunSetLogLevelCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{}
		},
	}

	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--profile", "p1", "--app", "api", "--log-level", "info"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunSetLogLevelCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--profile", "p1", "--app", "api", "--log-level", "info"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error setting log level") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunSetLogLevelCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogLevelDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runSetLogLevelCommand([]string{"--profile", "p1", "--app", "api", "--log-level", "info"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
