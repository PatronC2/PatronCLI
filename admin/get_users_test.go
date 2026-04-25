package admin

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunGetUsersCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotProfileName string
	var gotMethod string
	var gotURL string

	deps := getUsersDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "172.16.200.185", Port: "8443", Token: "Bearer x"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			return []byte(`{"data":[{"ID":1,"Username":"admin","PasswordHash":"","Role":"admin"},{"ID":7,"Username":"operator","PasswordHash":"","Role":"operator"}]}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetUsersCommand([]string{"--profile", "prod"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotProfileName != "prod" {
		t.Fatalf("expected profile 'prod', got %q", gotProfileName)
	}
	if gotMethod != "GET" {
		t.Fatalf("expected GET method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/users" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}

	output := out.String()
	if !strings.Contains(output, "\"Username\": \"admin\"") || !strings.Contains(output, "\"PasswordHash\": \"\"") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunGetUsersCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfileName string
	deps := getUsersDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"data":[]}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetUsersCommand([]string{}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfileName != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfileName)
	}
}

func TestRunGetUsersCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getUsersDeps{}
	var out bytes.Buffer
	err := runGetUsersCommand([]string{}, deps, &out)
	if !errors.Is(err, errGetUsersUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunGetUsersCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getUsersDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{}
		},
	}

	var out bytes.Buffer
	err := runGetUsersCommand([]string{"--profile", "p1"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunGetUsersCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getUsersDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runGetUsersCommand([]string{"--profile", "p1"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error fetching users") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunGetUsersCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := getUsersDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runGetUsersCommand([]string{"--profile", "p1"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
